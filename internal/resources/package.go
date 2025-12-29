package resources

import (
	"context"
	"fmt"
)

// PackageManager arayüzü: Farklı dağıtımlar için (Arch, Debian) ortak metodlar
type PackageManager interface {
	Install(name string) error
	Remove(name string) error
	Check(name string) (bool, error)
}

type PackageResource struct {
	CanonicalID string
	Name        string
	ManagerName string // "pacman", "apt" vb.
	State       string // "installed", "absent"
}

func (p *PackageResource) ID() string {
	if p.CanonicalID != "" {
		return p.CanonicalID
	}
	return fmt.Sprintf("package:%s", p.Name)
}

// getManager: Manager ismine göre uygun implementasyonu döner
func (p *PackageResource) getManager() PackageManager {
	switch p.ManagerName {
	case "pacman":
		return &ArchLinuxProvider{}
	// İleride buraya case "apt": return &DebianProvider{} eklenecek
	default:
		return &ArchLinuxProvider{} // Varsayılan
	}
}

func (p *PackageResource) Check() (bool, error) {
	mgr := p.getManager()
	exists, err := mgr.Check(p.Name)
	if err != nil {
		return false, err
	}

	if p.State == "absent" {
		return !exists, nil
	}
	return exists, nil
}

func (p *PackageResource) Apply() error {
	mgr := p.getManager()
	if p.State == "absent" {
		return mgr.Remove(p.Name)
	}
	return mgr.Install(p.Name)
}

func (p *PackageResource) Diff() (string, error) {
	exists, _ := p.getManager().Check(p.Name)
	if !exists && p.State != "absent" {
		return fmt.Sprintf("+ package: %s (Kurulacak)", p.Name), nil
	}
	if exists && p.State == "absent" {
		return fmt.Sprintf("- package: %s (Kaldırılacak)", p.Name), nil
	}
	return "", nil
}

func (p *PackageResource) Undo(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Undo mantığı: Eğer kurulduysa kaldır.
	return p.getManager().Remove(p.Name)
}
