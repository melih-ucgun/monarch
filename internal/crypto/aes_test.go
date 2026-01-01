package crypto

import (
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if len(key) != 64 { // 32 bytes * 2 hex chars
		t.Errorf("Expected key length 64, got %d", len(key))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := "my-super-secret-password"

	// Encrypt
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if !IsEncrypted(encrypted) {
		t.Error("IsEncrypted returned false for valid encrypted string")
	}

	// Decrypt
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text != plaintext. Got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptInvalidFormat(t *testing.T) {
	key, _ := GenerateKey()
	_, err := Decrypt("invalid-format", key)
	if err == nil {
		t.Error("Expected error for invalid format, got none")
	}
}

func TestDecryptInvalidKey(t *testing.T) {
	key, _ := GenerateKey()
	otherKey, _ := GenerateKey()
	plaintext := "secret"

	encrypted, _ := Encrypt(plaintext, key)

	// Try decrypting with wrong key
	_, err := Decrypt(encrypted, otherKey)
	if err == nil {
		t.Error("Expected error for wrong key, got none")
	}
}
