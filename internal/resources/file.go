package resources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type FileResource struct {
	CanonicalID string `mapstructure:"-"`
	Path        string `mapstructure:"path"`
	Content     string `mapstructure:"content"`
	Mode        string `mapstructure:"mode"` // Octal string örn: "0644"
}

func (r *FileResource) ID() string {
	return r.CanonicalID
}

func (r *FileResource) Check() (bool, error) {
	info, err := os.Stat(r.Path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("dosya kontrol hatası: %w", err)
	}

	// 1. İzin (Mode) Kontrolü
	currentPerm := info.Mode().Perm()
	expectedModeInt, err := strconv.ParseUint(r.Mode, 8, 32)
	if err != nil {
		return false, fmt.Errorf("geçersiz mode formatı (%s): %w", r.Mode, err)
	}

	if currentPerm != os.FileMode(expectedModeInt) {
		return false, nil
	}

	// 2. İçerik Kontrolü
	// Performans notu: Büyük dosyalarda hash kullanmak daha iyidir ama şimdilik direct read.
	content, err := os.ReadFile(r.Path)
	if err != nil {
		return false, fmt.Errorf("dosya okuma hatası: %w", err)
	}

	if string(content) != r.Content {
		return false, nil
	}

	return true, nil
}

func (r *FileResource) Apply() error {
	// Hedef dizinin varlığını garantiye al
	dir := filepath.Dir(r.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("hedef dizin oluşturulamadı: %w", err)
	}

	// Mode string'i os.FileMode'a çevir
	modeInt, err := strconv.ParseUint(r.Mode, 8, 32)
	if err != nil {
		return fmt.Errorf("mode parse hatası: %w", err)
	}
	mode := os.FileMode(modeInt)

	// Dosyayı yaz
	if err := os.WriteFile(r.Path, []byte(r.Content), mode); err != nil {
		return fmt.Errorf("dosya yazılamadı: %w", err)
	}

	// İzinleri garantiye al (WriteFile bazen umask'a takılabilir)
	if err := os.Chmod(r.Path, mode); err != nil {
		return fmt.Errorf("chmod hatası: %w", err)
	}

	return nil
}

func (r *FileResource) Undo(ctx context.Context) error {
	// Dikkat: Bu işlem dosyayı kalıcı olarak siler.
	return os.Remove(r.Path)
}

func (r *FileResource) Diff() (string, error) {
	// Basit bir diff mesajı
	return fmt.Sprintf("File[%s] content or permission mismatch", r.Path), nil
}
