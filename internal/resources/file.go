package resources

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// Apply ATOMİK çalışır ve CGO gerektirmez.
func (f *FileResource) Apply() error {
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("hedef dizin oluşturulamadı: %w", err)
	}

	// 1. Geçici dosya oluştur
	tmpFile, err := os.CreateTemp(dir, ".monarch-tmp-*")
	if err != nil {
		return fmt.Errorf("geçici dosya oluşturulamadı: %w", err)
	}
	tmpName := tmpFile.Name()

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

	// 3. Diske senkronize et (Sync)
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("disk senkronizasyonu hatası: %w", err)
	}

	// 4. Dosyayı kapat
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("dosya kapatılamadı: %w", err)
	}

	// 5. İzinleri ve Sahipliği ayarla
	if f.Mode != "" {
		m, _ := strconv.ParseUint(f.Mode, 8, 32)
		if err := os.Chmod(tmpName, os.FileMode(m)); err != nil {
			return fmt.Errorf("izinler ayarlanamadı: %w", err)
		}
	} else {
		os.Chmod(tmpName, 0o644)
	}

	if f.Owner != "" || f.Group != "" {
		uid, _ := resolveUser(f.Owner)
		gid, _ := resolveGroup(f.Group)
		// Not: Chown işlemi genellikle root yetkisi gerektirir.
		if err := os.Chown(tmpName, uid, gid); err != nil {
			return fmt.Errorf("sahiplik ayarlanamadı (root musunuz?): %w", err)
		}
	}

	// 6. Atomik Taşıma
	if err := os.Rename(tmpName, f.Path); err != nil {
		return fmt.Errorf("atomik taşıma başarısız: %w", err)
	}

	success = true
	return nil
}

// Undo, dosya kaynağını sistemden kaldırır (Detach işlemi).
func (f *FileResource) Undo(ctx context.Context) error {
	// 1. Dosya var mı kontrol et
	if _, err := os.Stat(f.Path); os.IsNotExist(err) {
		// Zaten yoksa işlem başarılıdır (Idempotency)
		return nil
	}

	// 2. Context iptal edildi mi kontrol et
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 3. Dosyayı sil
	if err := os.Remove(f.Path); err != nil {
		return fmt.Errorf("dosya silinemedi [%s]: %w", f.Path, err)
	}

	return nil
}

// resolveUser: /etc/passwd dosyasını okuyarak kullanıcı adını UID'ye çevirir.
// CGO (os/user) kullanmaz, böylece statik binary olarak derlenebilir.
func resolveUser(name string) (int, error) {
	if name == "" {
		return -1, nil
	}

	// Eğer input zaten sayıysa (örn: "1000"), direkt döndür.
	if id, err := strconv.Atoi(name); err == nil {
		return id, nil
	}

	f, err := os.Open("/etc/passwd")
	if err != nil {
		return -1, fmt.Errorf("passwd dosyası okunamadı: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Yorum satırlarını veya boş satırları atla
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Format: username:password:uid:gid:gecos:home:shell
		parts := strings.Split(line, ":")
		if len(parts) > 2 && parts[0] == name {
			uid, err := strconv.Atoi(parts[2])
			if err != nil {
				return -1, fmt.Errorf("geçersiz uid formatı: %w", err)
			}
			return uid, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return -1, err
	}

	return -1, fmt.Errorf("kullanıcı bulunamadı: %s", name)
}

// resolveGroup: /etc/group dosyasını okuyarak grup adını GID'ye çevirir.
// CGO kullanmaz.
func resolveGroup(name string) (int, error) {
	if name == "" {
		return -1, nil
	}

	if id, err := strconv.Atoi(name); err == nil {
		return id, nil
	}

	f, err := os.Open("/etc/group")
	if err != nil {
		return -1, fmt.Errorf("group dosyası okunamadı: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Format: groupname:password:gid:userlist
		parts := strings.Split(line, ":")
		if len(parts) > 2 && parts[0] == name {
			gid, err := strconv.Atoi(parts[2])
			if err != nil {
				return -1, fmt.Errorf("geçersiz gid formatı: %w", err)
			}
			return gid, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return -1, err
	}

	return -1, fmt.Errorf("grup bulunamadı: %s", name)
}
