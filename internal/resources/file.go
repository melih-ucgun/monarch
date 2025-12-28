package resources

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type FileResource struct {
	CanonicalID  string // Factory tarafından atanan kimlik
	ResourceName string
	Path         string
	Content      string
}

func (f *FileResource) ID() string {
	return f.CanonicalID
}

func (f *FileResource) Check() (bool, error) {
	info, err := os.Stat(f.Path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("dosya kontrol edilemedi: %w", err)
	}

	if info.IsDir() {
		return false, fmt.Errorf("%s bir dizin, dosya olması bekleniyordu", f.Path)
	}

	currentContent, err := os.ReadFile(f.Path)
	if err != nil {
		return false, fmt.Errorf("dosya okunamadı: %w", err)
	}

	return bytes.Equal(currentContent, []byte(f.Content)), nil
}

func (f *FileResource) Apply() error {
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("dizin oluşturulamadı %s: %w", dir, err)
	}

	err := os.WriteFile(f.Path, []byte(f.Content), 0o644)
	if err != nil {
		return fmt.Errorf("dosya yazılamadı %s: %w", f.Path, err)
	}

	return nil
}
