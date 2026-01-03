package transport

import (
	"context"
	"testing"
	"time"
)

func TestNewSSHTransport(t *testing.T) {
	// HostName alanı Name olarak güncellendi
	host := HostConfig{
		Name:           "test-host",
		Address:        "127.0.0.1",
		User:           "testuser",
		Port:           22,
		BecomePassword: "password",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Bağlantı denemesi (bağlanamayacağı için hata vermesi normal,
	// ancak struct literal hatası vermemeli)
	_, err := NewSSHTransport(ctx, host)
	if err == nil {
		t.Log("Beklenen: Bağlantı hatası aldık (sunucu yok)")
	}
}
