package resources

import (
	"testing"

	"github.com/melih-ucgun/monarch/internal/config"
)

func TestNewFileResource(t *testing.T) {
	cfg := config.ResourceConfig{
		ID:   "file:test",
		Type: "file",
		Parameters: map[string]interface{}{
			"path":    "/tmp/test",
			"content": "hello",
			"mode":    "0644",
		},
	}

	res, err := New(cfg, nil)
	if err != nil {
		t.Fatalf("Factory hata döndü: %v", err)
	}

	fileRes, ok := res.(*FileResource)
	if !ok {
		t.Fatalf("Dönen tip FileResource değil")
	}

	if fileRes.Path != "/tmp/test" {
		t.Errorf("Path hatalı: %s", fileRes.Path)
	}

	// Interface uyumluluk testi
	var _ Resource = fileRes
}

func TestNewInvalidResource(t *testing.T) {
	cfg := config.ResourceConfig{
		Type: "unknown_type",
	}

	_, err := New(cfg, nil)
	if err == nil {
		t.Error("Bilinmeyen tip için hata dönmeliydi")
	}
}
