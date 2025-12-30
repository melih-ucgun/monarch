package resources

import (
	"context"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

type UserResource struct {
	CanonicalID string   `mapstructure:"-"`
	Name        string   `mapstructure:"name"`
	Groups      []string `mapstructure:"groups"` // Ekstra gruplar (append)
	Shell       string   `mapstructure:"shell"`
	System      bool     `mapstructure:"system"` // Sistem kullanıcısı mı?
}

func (r *UserResource) ID() string {
	return r.CanonicalID
}

func (r *UserResource) Check() (bool, error) {
	// Kullanıcı varlık kontrolü (u değişkeni kullanılmıyor, _ yapıldı)
	_, err := user.Lookup(r.Name)
	if err != nil {
		// Kullanıcı yoksa oluşturulmalı -> Check başarısız
		return false, nil
	}

	// Grupları kontrol et
	// `id -Gn user` komutu kullanıcının gruplarını isim olarak döner
	cmd := exec.Command("id", "-Gn", r.Name)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("grup kontrol hatası: %w", err)
	}
	currentGroups := strings.Fields(string(out))

	// İstenen her grup mevcut mu?
	for _, requiredGroup := range r.Groups {
		found := false
		for _, g := range currentGroups {
			if g == requiredGroup {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	// Shell kontrolü burada yapılabilir ancak basitlik adına atlıyoruz.
	// Eğer shell değişmişse Apply adımında usermod ile düzeltilecek.

	return true, nil
}

func (r *UserResource) Apply() error {
	// Kullanıcı var mı?
	_, err := user.Lookup(r.Name)

	if err != nil {
		// Kullanıcı YOK, oluştur
		args := []string{"useradd", "-m", r.Name}
		if r.Shell != "" {
			args = append(args, "-s", r.Shell)
		}
		if r.System {
			args = append(args, "-r")
		}
		if len(r.Groups) > 0 {
			args = append(args, "-G", strings.Join(r.Groups, ","))
		}

		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("useradd hatası: %s", string(out))
		}
	} else {
		// Kullanıcı VAR, güncelle (usermod)
		args := []string{"usermod"}

		if len(r.Groups) > 0 {
			// -aG: Append groups
			args = append(args, "-aG", strings.Join(r.Groups, ","))
		}
		if r.Shell != "" {
			args = append(args, "-s", r.Shell)
		}

		// Sadece usermod komutu değilse (argüman eklendiyse) çalıştır
		if len(args) > 1 {
			args = append(args, r.Name)
			if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
				return fmt.Errorf("usermod hatası: %s", string(out))
			}
		}
	}

	return nil
}

func (r *UserResource) Undo(ctx context.Context) error {
	// Kullanıcı silmek tehlikeli olabilir.
	return exec.Command("userdel", r.Name).Run()
}

func (r *UserResource) Diff() (string, error) {
	return fmt.Sprintf("User[%s] missing or config mismatch", r.Name), nil
}
