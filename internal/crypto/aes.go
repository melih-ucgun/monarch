package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	Prefix = "ENC[AES256:"
	Suffix = "]"
)

// GenerateKey generates a random 32-byte key and returns it as a hex string.
func GenerateKey() (string, error) {
	bytes := make([]byte, 32) // AES-256
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Encrypt encrypts plaintext using AES-GCM with the provided hex-encoded key.
// Returns formatted string: ENC[AES256:<base64_ciphertext>]
func Encrypt(plaintext, keyHex string) (string, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("%s%s%s", Prefix, base64.StdEncoding.EncodeToString(ciphertext), Suffix), nil
}

// Decrypt decrypts a formatted string using the provided hex-encoded key.
// Expected format: ENC[AES256:<base64_ciphertext>]
func Decrypt(encrypted, keyHex string) (string, error) {
	if !strings.HasPrefix(encrypted, Prefix) || !strings.HasSuffix(encrypted, Suffix) {
		return "", errors.New("invalid encrypted format")
	}

	// Remove Prefix and Suffix
	b64 := encrypted[len(Prefix) : len(encrypted)-len(Suffix)]
	ciphertext, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a string follows the encrypted format.
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, Prefix) && strings.HasSuffix(s, Suffix)
}
