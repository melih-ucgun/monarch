package resources

import (
	"context"
	"fmt"
	"os/exec"
	"os/user"
)

type GroupResource struct {
	CanonicalID string `mapstructure:"-"`
	Name        string `mapstructure:"name"`
	GID         string `mapstructure:"gid"`    // Opsiyonel
	System      bool   `mapstructure:"system"` // Sistem grubu mu?
}

func (r *GroupResource) ID() string {
	return r.CanonicalID
}

func (r *GroupResource) Check() (bool, error) {
	// Grubu adına göre ara
	_, err := user.LookupGroup(r.Name)
	if err != nil {
		// Grup yok
		return false, nil
	}

	// GID kontrolü (Eğer belirtilmişse)
	// user.LookupGroup Gid string döner, karşılaştırılabilir.
	// Ancak basitlik adına şimdilik sadece varlık kontrolü yapıyoruz.
	return true, nil
}

func (r *GroupResource) Apply() error {
	// Grup var mı tekrar kontrol et (Lookup maliyetli değildir)
	_, err := user.LookupGroup(r.Name)
	if err == nil {
		// Grup zaten var, güncelleme gerekirse 'groupmod' kullanılabilir
		return nil
	}

	// Grup yok, oluştur
	args := []string{"groupadd"}
	if r.GID != "" {
		args = append(args, "-g", r.GID)
	}
	if r.System {
		args = append(args, "-r")
	}
	args = append(args, r.Name)

	if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
		return fmt.Errorf("groupadd hatası: %s", string(out))
	}

	return nil
}

func (r *GroupResource) Undo(ctx context.Context) error {
	return exec.Command("groupdel", r.Name).Run()
}

func (r *GroupResource) Diff() (string, error) {
	return fmt.Sprintf("Group[%s] missing", r.Name), nil
}
