package resources

import (
	"fmt"
)

// PackageManager, farklı paket yöneticileri için ortak arayüzdür.
type PackageManager interface {
	IsInstalled(name string) (bool, error)
	Install(name string) error
	Remove(name string) error
}

// GetPackageManager, istenen isme göre uygun paket yöneticisini döndürür.
func GetPackageManager(managerName string) (PackageManager, error) {
	switch managerName {
	// --- Sistem Paket Yöneticileri ---
	case "pacman":
		return &PacmanManager{}, nil
	case "apt":
		return &AptManager{}, nil
	case "dnf":
		return &DnfManager{}, nil
	case "zypper":
		return &ZypperManager{}, nil
	case "yum":
		return &YumManager{}, nil
	case "apk":
		return &ApkManager{}, nil
	case "brew":
		return &BrewManager{}, nil

	// --- AUR Yardımcıları ---
	case "paru":
		return &ParuManager{}, nil
	case "yay":
		return &YayManager{}, nil

	// --- Evrensel Paket Formatları ---
	case "flatpak":
		return &FlatpakManager{}, nil
	case "snap":
		return &SnapManager{}, nil

	default:
		return nil, fmt.Errorf("desteklenmeyen paket yöneticisi: %s", managerName)
	}
}
