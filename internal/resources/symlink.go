package resources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type SymlinkResource struct {
	CanonicalID string `mapstructure:"-"`
	Target      string `mapstructure:"target"` // Gerçek dosya
	Link        string `mapstructure:"link"`   // Oluşturulacak link
	Force       bool   `mapstructure:"force"`
}

func (r *SymlinkResource) ID() string {
	return r.CanonicalID
}

func (r *SymlinkResource) Check() (bool, error) {
	info, err := os.Lstat(r.Link)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Link mi?
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil // Var ama link değil
	}

	// Hedef doğru mu?
	dest, err := os.Readlink(r.Link)
	if err != nil {
		return false, err
	}

	return dest == r.Target, nil
}

func (r *SymlinkResource) Apply() error {
	// Hedef dizini oluştur
	dir := filepath.Dir(r.Link)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("parent dizin oluşturulamadı: %w", err)
	}

	// Eğer dosya/link varsa ve force ise veya yanlışsa sil
	if _, err := os.Lstat(r.Link); err == nil {
		if r.Force {
			if err := os.Remove(r.Link); err != nil {
				return fmt.Errorf("eski link silinemedi: %w", err)
			}
		} else {
			// Force değilse ve çakışma varsa hata dönmek daha güvenli olabilir
			// ama idempotent olması için burada kontrol edip sadece yanlışsa silebiliriz.
			// Şimdilik force yoksa dokunmuyoruz (Check zaten false dönerse Apply çalışır)
			// Apply çağrıldıysa ve dosya varsa, demek ki Check false döndü.
			// Yani dosya yanlış. Silip tekrar yapmalıyız.
			if err := os.Remove(r.Link); err != nil {
				return fmt.Errorf("mevcut dosya silinemedi (override): %w", err)
			}
		}
	}

	if err := os.Symlink(r.Target, r.Link); err != nil {
		return fmt.Errorf("symlink oluşturulamadı: %w", err)
	}
	return nil
}

func (r *SymlinkResource) Undo(ctx context.Context) error {
	return os.Remove(r.Link)
}

func (r *SymlinkResource) Diff() (string, error) {
	return fmt.Sprintf("Symlink[%s -> %s] mismatch", r.Link, r.Target), nil
}
