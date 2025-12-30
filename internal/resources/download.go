package resources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type DownloadResource struct {
	CanonicalID string `mapstructure:"-"`
	URL         string `mapstructure:"url"`
	Dest        string `mapstructure:"dest"`
	Mode        string `mapstructure:"mode"` // Örn: "0755"
}

func (r *DownloadResource) ID() string {
	return r.CanonicalID
}

func (r *DownloadResource) Check() (bool, error) {
	// 1. Dosya var mı?
	info, err := os.Stat(r.Dest)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// 2. Mode kontrolü (Opsiyonel ama iyi olur)
	if r.Mode != "" {
		expectedModeInt, err := strconv.ParseUint(r.Mode, 8, 32)
		if err == nil {
			if info.Mode().Perm() != os.FileMode(expectedModeInt) {
				return false, nil
			}
		}
	}

	// İleri seviye: Checksum kontrolü (md5/sha256) eklenebilir.
	// Şimdilik sadece varlık kontrolü yapıyoruz.
	return true, nil
}

func (r *DownloadResource) Apply() error {
	// Hedef dizini oluştur
	if err := os.MkdirAll(filepath.Dir(r.Dest), 0755); err != nil {
		return fmt.Errorf("hedef dizin oluşturulamadı: %w", err)
	}

	// İndirme işlemini başlat
	resp, err := http.Get(r.URL)
	if err != nil {
		return fmt.Errorf("indirme başarısız (%s): %w", r.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sunucu hatası: %s", resp.Status)
	}

	// Dosyayı oluştur
	out, err := os.Create(r.Dest)
	if err != nil {
		return fmt.Errorf("dosya oluşturulamadı: %w", err)
	}
	defer out.Close()

	// İçeriği kopyala
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("yazma hatası: %w", err)
	}

	// İzinleri ayarla
	if r.Mode != "" {
		modeInt, err := strconv.ParseUint(r.Mode, 8, 32)
		if err == nil {
			if err := os.Chmod(r.Dest, os.FileMode(modeInt)); err != nil {
				return fmt.Errorf("chmod hatası: %w", err)
			}
		}
	}

	return nil
}

func (r *DownloadResource) Undo(ctx context.Context) error {
	return os.Remove(r.Dest)
}

func (r *DownloadResource) Diff() (string, error) {
	return fmt.Sprintf("Download[%s] from %s missing", r.Dest, r.URL), nil
}
