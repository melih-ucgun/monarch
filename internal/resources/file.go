package resources

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

type FileResource struct {
	CanonicalID  string
	ResourceName string
	Path         string
	Content      string
	Mode         string
	Owner        string
	Group        string
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
		return false, err
	}

	// 1. İçerik
	currentContent, err := os.ReadFile(f.Path)
	if err != nil {
		return false, err
	}
	if !bytes.Equal(currentContent, []byte(f.Content)) {
		return false, nil
	}

	// 2. İzinler
	if f.Mode != "" {
		targetMode, _ := strconv.ParseUint(f.Mode, 8, 32)
		if uint32(info.Mode().Perm()) != uint32(targetMode) {
			return false, nil
		}
	}

	// 3. Sahiplik
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if f.Owner != "" {
			uid, _ := resolveUser(f.Owner)
			if uid != -1 && stat.Uid != uint32(uid) {
				return false, nil
			}
		}
		if f.Group != "" {
			gid, _ := resolveGroup(f.Group)
			if gid != -1 && stat.Gid != uint32(gid) {
				return false, nil
			}
		}
	}

	return true, nil
}

func (f *FileResource) Diff() (string, error) {
	info, err := os.Stat(f.Path)
	if os.IsNotExist(err) {
		return fmt.Sprintf("+ file: %s (Oluşturulacak)", f.Path), nil
	}

	diffMsg := ""
	current, _ := os.ReadFile(f.Path)
	if string(current) != f.Content {
		diffMsg += "~ İçerik değişecek\n"
	}

	if f.Mode != "" {
		m, _ := strconv.ParseUint(f.Mode, 8, 32)
		if uint32(info.Mode().Perm()) != uint32(m) {
			diffMsg += fmt.Sprintf("~ İzinler: %o -> %s\n", info.Mode().Perm(), f.Mode)
		}
	}

	if diffMsg == "" {
		return "", nil
	}
	return fmt.Sprintf("! %s:\n%s", f.Path, diffMsg), nil
}

// Apply artık ATOMİK çalışır.
// Dosyayı önce geçici bir isme yazar, sonra rename yapar.
func (f *FileResource) Apply() error {
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("hedef dizin oluşturulamadı: %w", err)
	}

	// 1. Geçici dosya oluştur (Aynı dizinde olmalı ki rename işlemi atomik olsun)
	tmpFile, err := os.CreateTemp(dir, ".monarch-tmp-*")
	if err != nil {
		return fmt.Errorf("geçici dosya oluşturulamadı: %w", err)
	}
	tmpName := tmpFile.Name()

	// İşlem başarısız olursa geçici dosyayı temizle
	success := false
	defer func() {
		tmpFile.Close() // Dosyayı kapatmayı garantiye al
		if !success {
			os.Remove(tmpName) // Başarısızsa sil
		}
	}()

	// 2. İçeriği yaz
	if _, err := tmpFile.Write([]byte(f.Content)); err != nil {
		return fmt.Errorf("dosya içeriği yazılamadı: %w", err)
	}

	// 3. Diske senkronize et (Veri bütünlüğü için kritik)
	// Bu işlem verinin işletim sistemi tamponundan diske fiziksel olarak yazılmasını zorlar.
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("disk senkronizasyonu hatası: %w", err)
	}

	// 4. Dosyayı kapat (Rename etmeden önce kapatmak şarttır)
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("dosya kapatılamadı: %w", err)
	}

	// 5. İzinleri ve Sahipliği ayarla (Rename öncesi yapılmalı)
	// Böylece dosya görünür olduğunda zaten doğru izinlere sahip olur.
	if f.Mode != "" {
		m, _ := strconv.ParseUint(f.Mode, 8, 32)
		if err := os.Chmod(tmpName, os.FileMode(m)); err != nil {
			return fmt.Errorf("izinler ayarlanamadı: %w", err)
		}
	} else {
		// Varsayılan izin (orijinal kodda olduğu gibi)
		os.Chmod(tmpName, 0o644)
	}

	if f.Owner != "" || f.Group != "" {
		uid, _ := resolveUser(f.Owner)
		gid, _ := resolveGroup(f.Group)
		if err := os.Chown(tmpName, uid, gid); err != nil {
			return fmt.Errorf("sahiplik ayarlanamadı: %w", err)
		}
	}

	// 6. Atomik Taşıma (Atomic Rename)
	// Eski dosya varsa güvenli bir şekilde ezilir.
	if err := os.Rename(tmpName, f.Path); err != nil {
		return fmt.Errorf("atomik taşıma başarısız: %w", err)
	}

	success = true
	return nil
}

func resolveUser(name string) (int, error) {
	if name == "" {
		return -1, nil
	}
	u, err := user.Lookup(name)
	if err != nil {
		if id, errID := strconv.Atoi(name); errID == nil {
			return id, nil
		}
		return -1, err
	}
	uid, _ := strconv.Atoi(u.Uid)
	return uid, nil
}

func resolveGroup(name string) (int, error) {
	if name == "" {
		return -1, nil
	}
	g, err := user.LookupGroup(name)
	if err != nil {
		if id, errID := strconv.Atoi(name); errID == nil {
			return id, nil
		}
		return -1, err
	}
	gid, _ := strconv.Atoi(g.Gid)
	return gid, nil
}
