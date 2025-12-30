package resources

import (
	"context"
	"fmt"
)

type PackageResource struct {
	CanonicalID string `mapstructure:"-"`
	Name        string `mapstructure:"name"`
	ManagerName string `mapstructure:"manager"` // pacman, apt, brew vs.
	State       string `mapstructure:"state"`   // installed, removed
}

func (r *PackageResource) ID() string {
	return r.CanonicalID
}

func (r *PackageResource) Check() (bool, error) {
	// package_arch.go içerisindeki GetPackageManager'ı çağırır
	mgr, err := GetPackageManager(r.ManagerName)
	if err != nil {
		return false, fmt.Errorf("paket yöneticisi hatası: %w", err)
	}

	isInstalled, err := mgr.IsInstalled(r.Name)
	if err != nil {
		return false, fmt.Errorf("paket durumu sorgulanamadı (%s): %w", r.Name, err)
	}

	switch r.State {
	case "installed":
		return isInstalled, nil
	case "removed":
		return !isInstalled, nil
	default:
		return false, fmt.Errorf("desteklenmeyen paket durumu: %s", r.State)
	}
}

func (r *PackageResource) Apply() error {
	mgr, err := GetPackageManager(r.ManagerName)
	if err != nil {
		return err
	}

	switch r.State {
	case "installed":
		if err := mgr.Install(r.Name); err != nil {
			return fmt.Errorf("paket yüklenemedi (%s): %w", r.Name, err)
		}
	case "removed":
		if err := mgr.Remove(r.Name); err != nil {
			return fmt.Errorf("paket kaldırılamadı (%s): %w", r.Name, err)
		}
	}
	return nil
}

func (r *PackageResource) Undo(ctx context.Context) error {
	mgr, err := GetPackageManager(r.ManagerName)
	if err != nil {
		return err
	}

	// Basit undo mantığı: Yüklendiyse kaldır, kaldırıldıysa yükle
	if r.State == "installed" {
		return mgr.Remove(r.Name)
	} else if r.State == "removed" {
		return mgr.Install(r.Name)
	}
	return nil
}

func (r *PackageResource) Diff() (string, error) {
	return fmt.Sprintf("Package[%s] state mismatch: want %s", r.Name, r.State), nil
}
