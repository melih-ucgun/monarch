package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/melih-ucgun/monarch/internal/core"
)

type ArchiveAdapter struct {
	core.BaseResource
	Source string // Arşiv dosyasının yolu (örn: /tmp/app.zip)
	Dest   string // Nereye açılacağı (örn: /opt/app)
	Mode   os.FileMode
}

func NewArchiveAdapter(name string, params map[string]interface{}) *ArchiveAdapter {
	src, _ := params["source"].(string)
	if src == "" {
		src = name
	} // Eğer source verilmezse isim source olarak kullanılır

	dest, _ := params["dest"].(string)
	if dest == "" {
		// Dest verilmezse source ile aynı yere klasör olarak aç
		dest = strings.TrimSuffix(src, filepath.Ext(src))
	}

	mode := os.FileMode(0755)
	if m, ok := params["mode"].(int); ok {
		mode = os.FileMode(m)
	}

	return &ArchiveAdapter{
		BaseResource: core.BaseResource{Name: name, Type: "archive"},
		Source:       src,
		Dest:         dest,
		Mode:         mode,
	}
}

func (r *ArchiveAdapter) Validate() error {
	if r.Source == "" {
		return fmt.Errorf("archive source is required")
	}
	if r.Dest == "" {
		return fmt.Errorf("archive destination is required")
	}
	return nil
}

func (r *ArchiveAdapter) Check(ctx *core.SystemContext) (bool, error) {
	// 1. Kaynak dosya var mı?
	if _, err := os.Stat(r.Source); os.IsNotExist(err) {
		return false, fmt.Errorf("archive source not found: %s", r.Source)
	}

	// 2. Hedef klasör var mı?
	// Basit mantık: Hedef klasör yoksa arşiv açılmalı -> Değişiklik var (true)
	// Daha gelişmiş mantık: Hedef klasör boş mu? İçinde belirli bir dosya var mı?
	// Şimdilik sadece klasör varlığına bakıyoruz.
	if _, err := os.Stat(r.Dest); os.IsNotExist(err) {
		return true, nil
	}

	// Hedef zaten varsa değişiklik yok varsayıyoruz (Force update logic'i eklenebilir)
	return false, nil
}

func (r *ArchiveAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, err := r.Check(ctx)
	if err != nil {
		return core.Failure(err, "Check failed"), err
	}
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Archive already extracted to %s", r.Dest)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] Extract %s to %s", r.Source, r.Dest)), nil
	}

	// Hedef klasörü oluştur
	if err := os.MkdirAll(r.Dest, r.Mode); err != nil {
		return core.Failure(err, "Failed to create destination directory"), err
	}

	// Dosya uzantısına göre açma işlemi
	if strings.HasSuffix(r.Source, ".zip") {
		err = unzip(r.Source, r.Dest)
	} else if strings.HasSuffix(r.Source, ".tar.gz") || strings.HasSuffix(r.Source, ".tgz") {
		err = untar(r.Source, r.Dest)
	} else {
		return core.Failure(nil, "Unsupported archive format"), fmt.Errorf("unsupported format: %s", r.Source)
	}

	if err != nil {
		return core.Failure(err, "Extraction failed"), err
	}

	return core.SuccessChange(fmt.Sprintf("Archive extracted to %s", r.Dest)), nil
}

// --- Yardımcı Fonksiyonlar ---

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Zip Slip zafiyetini önle
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		// Zip Slip önlemi
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}
