package resources

import "os/exec"

type PackageManager interface {
	IsInstalled(name string) (bool, error)
	Install(name string) error
	Remove(name string) error
}

type PackageResource struct {
	PackageName string
	State       string
	Provider    PackageManager
}

// ID metodunu panic yerine gerçek bir ID dönecek şekilde düzeltiyoruz
func (p *PackageResource) ID() string {
	return "pkg:" + p.PackageName
}

func (p *PackageResource) Check() (bool, error) {
	return p.Provider.IsInstalled(p.PackageName)
}

func (p *PackageResource) Apply() error {
	if p.State == "installed" || p.State == "" {
		return p.Provider.Install(p.PackageName)
	}
	return nil
}

// --- OTOMATİK ALGILAMA MEKANİZMASI ---
func GetDefaultProvider() PackageManager {
	// 1. Sistemde pacman var mı kontrol et (CachyOS/Arch için)
	if _, err := exec.LookPath("pacman"); err == nil {
		return &PacmanProvider{}
	}

	// 2. Sistemde apt var mı kontrol et (Debian/Ubuntu için)
	if _, err := exec.LookPath("apt-get"); err == nil {
		// return &AptProvider{} // İleride buraya AptProvider eklenebilir
	}

	return nil
}
