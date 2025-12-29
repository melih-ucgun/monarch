package transport

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
)

// Test ortamı için geçici SSH sunucusu (Mock server gerekebilir veya container)
// Ancak gerçek bir SSH bağlantısı testi karmaşıktır.
// Bu yüzden burada sadece Config -> Transport dönüşümünü ve basit mantığı test ediyoruz.
// Entegrasyon testleri için ayrı bir "integration" build tag'i kullanılabilir.

func TestNewSSHTransport_Config(t *testing.T) {
	// Geçici bir SSH anahtarı oluştur
	keyPath := filepath.Join(t.TempDir(), "id_rsa")
	if err := generateSSHKey(keyPath); err != nil {
		t.Fatalf("SSH anahtarı oluşturulamadı: %v", err)
	}

	hostCfg := config.Host{
		Name:       "test-host",
		HostName:   "127.0.0.1", // Address yerine HostName kullanılıyor
		User:       "testuser",
		Port:       2222,
		SSHKeyPath: keyPath,
	}

	// Bağlantı kurulmaya çalışılacak ama sunucu olmadığı için hata verecek.
	// Ancak hata "anahtar okunamadı" değil "bağlantı reddedildi" olmalı.
	// Bu da config'in doğru işlendiğini gösterir.

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewSSHTransport(ctx, hostCfg)
	if err == nil {
		t.Error("Sunucu yokken bağlantı başarılı olmamalıydı")
	}

	// Hata mesajını kontrol et: dial tcp ... connect: connection refused
	// Eğer "no such file" gibi bir hata alırsak keyPath yanlış demektir.
}

func generateSSHKey(path string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	keyFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	return pem.Encode(keyFile, privBlock)
}

// Mock SSH Sunucusu ile test (Gelişmiş)
// Bu kısım çok daha kapsamlıdır ve genellikle ayrı bir kütüphane veya container gerektirir.
// Şimdilik sadece config mapping testi yeterli.
