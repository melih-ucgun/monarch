package resources

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type TemplateResource struct {
	CanonicalID string            `mapstructure:"-"`
	Src         string            `mapstructure:"src"`  // Şablon dosyasının yolu
	Dest        string            `mapstructure:"dest"` // Hedef dosya yolu
	Mode        os.FileMode       `mapstructure:"-"`    // Factory'de manuel set edilir
	Vars        map[string]string `mapstructure:"-"`    // Global değişkenler
}

func (r *TemplateResource) ID() string {
	return r.CanonicalID
}

// render fonksiyonu şablonu bellekte işler
func (r *TemplateResource) render() ([]byte, error) {
	// Şablon dosyasını oku
	tmplContent, err := os.ReadFile(r.Src)
	if err != nil {
		return nil, fmt.Errorf("şablon okunamadı (%s): %w", r.Src, err)
	}

	// Go template motorunu hazırla
	tmpl, err := template.New(filepath.Base(r.Src)).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("şablon parse hatası: %w", err)
	}

	// Değişkenleri uygula (Execute)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.Vars); err != nil {
		return nil, fmt.Errorf("şablon render hatası: %w", err)
	}

	return buf.Bytes(), nil
}

func (r *TemplateResource) Check() (bool, error) {
	// 1. Hedef dosya var mı?
	info, err := os.Stat(r.Dest)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// 2. İçerik kontrolü
	// Beklenen içeriği render et
	expectedContent, err := r.render()
	if err != nil {
		return false, err
	}

	// Mevcut içeriği oku
	currentContent, err := os.ReadFile(r.Dest)
	if err != nil {
		return false, err
	}

	// Byte karşılaştırması
	if !bytes.Equal(currentContent, expectedContent) {
		return false, nil
	}

	// 3. Mode kontrolü
	if info.Mode().Perm() != r.Mode {
		return false, nil
	}

	return true, nil
}

func (r *TemplateResource) Apply() error {
	// İçeriği oluştur
	content, err := r.render()
	if err != nil {
		return err
	}

	// Hedef dizini garantiye al
	if err := os.MkdirAll(filepath.Dir(r.Dest), 0755); err != nil {
		return fmt.Errorf("hedef dizin oluşturulamadı: %w", err)
	}

	// Dosyayı yaz
	if err := os.WriteFile(r.Dest, content, r.Mode); err != nil {
		return fmt.Errorf("dosya yazılamadı: %w", err)
	}

	// İzinleri garantiye al
	if err := os.Chmod(r.Dest, r.Mode); err != nil {
		return fmt.Errorf("chmod hatası: %w", err)
	}

	return nil
}

func (r *TemplateResource) Undo(ctx context.Context) error {
	return os.Remove(r.Dest)
}

func (r *TemplateResource) Diff() (string, error) {
	return fmt.Sprintf("Template[%s -> %s] content mismatch", r.Src, r.Dest), nil
}
