package resources

import (
	"fmt"
	"os/exec"
)

type PackageResource struct {
	PackageName string
	State       string // "installed" veya "absent"
}

func (p *PackageResource) ID() string {
	return fmt.Sprintf("pkg:%s", p.PackageName)
}

// Check, paketin kurulu olup olmadığını kontrol eder.
func (p *PackageResource) Check() (bool, error) {
	// Arch Linux (CachyOS) için pacman -Q komutu kullanılır.
	cmd := exec.Command("pacman", "-Q", p.PackageName)
	err := cmd.Run()
	if err != nil {
		// Paket kurulu değilse pacman hata kodu döndürür.
		return false, nil
	}
	return true, nil
}

// Apply, paketi kurar.
func (p *PackageResource) Apply() error {
	fmt.Printf("📦 Installing package: %s...\n", p.PackageName)

	// -S: Kur, --noconfirm: Onay sormadan devam et.
	// NOT: Bu işlem genellikle sudo yetkisi gerektirir.
	cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", p.PackageName)

	// Çıktıyı terminalde görmek istersen:
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("paket kurulumu başarısız: %w", err)
	}
	return nil
}
