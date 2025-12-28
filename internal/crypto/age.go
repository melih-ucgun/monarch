package crypto

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
)

// Decrypt, age ile şifrelenmiş bir metni verilen private key ile çözer.
func Decrypt(encryptedData string, privateKey string) (string, error) {
	// Private key'i ayrıştır
	identity, err := age.ParseIdentities(strings.NewReader(privateKey))
	if err != nil {
		return "", fmt.Errorf("private key okunamadı: %w", err)
	}

	// Şifreli metni oku (Base64 veya direkt metin formatında gelebilir, age genellikle ASCII Armor kullanır)
	r, err := age.Decrypt(strings.NewReader(encryptedData), identity...)
	if err != nil {
		return "", fmt.Errorf("deşifreleme başarısız: %w", err)
	}

	out := &bytes.Buffer{}
	if _, err := io.Copy(out, r); err != nil {
		return "", fmt.Errorf("çıktı kopyalanamadı: %w", err)
	}

	return out.String(), nil
}

// GenerateKey, yeni bir age anahtar çifti oluşturur.
func GenerateKey() (string, string, error) {
	k, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", err
	}
	return k.String(), k.Recipient().String(), nil
}

// Encrypt, bir metni verilen public key (recipient) ile şifreler.
func Encrypt(plainText string, publicKey string) (string, error) {
	recipient, err := age.ParseRecipients(strings.NewReader(publicKey))
	if err != nil {
		return "", err
	}

	out := &bytes.Buffer{}
	w, err := age.Encrypt(out, recipient...)
	if err != nil {
		return "", err
	}

	if _, err := io.WriteString(w, plainText); err != nil {
		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	return out.String(), nil
}
