package resources

import (
	"fmt"
	"os/exec"
)

type ArchLinuxProvider struct{}

// detectHelper: Sistemdeki AUR yardımcısını (paru veya yay) bulur.
// Öncelik Paru'dadır. Hiçbiri yoksa boş string döner.
func (a *ArchLinuxProvider) detectHelper() string {
	if _, err := exec.LookPath("paru"); err == nil {
		return "paru"
	}
	if _, err := exec.LookPath("yay"); err == nil {
		return "yay"
	}
	return ""
}

func (a *ArchLinuxProvider) Install(name string) error {
	helper := a.detectHelper()

	var cmd *exec.Cmd
	if helper != "" {
		fmt.Printf("📦 Paket kuruluyor (%s): %s\n", helper, name)
		// AUR yardımcıları (paru/yay) genellikle sudo ile çalıştırılmaz,
		// root yetkisini kendileri isterler.
		cmd = exec.Command(helper, "-S", "--noconfirm", "--needed", name)
	} else {
		fmt.Printf("📦 Paket kuruluyor (Pacman): %s\n", name)
		// Pacman sudo gerektirir
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "--needed", name)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s install hatası: %s", helper, string(out))
	}
	return nil
}

func (a *ArchLinuxProvider) Remove(name string) error {
	helper := a.detectHelper()

	var cmd *exec.Cmd
	if helper != "" {
		fmt.Printf("🗑️ Paket siliniyor (%s): %s\n", helper, name)
		cmd = exec.Command(helper, "-Rns", "--noconfirm", name)
	} else {
		fmt.Printf("🗑️ Paket siliniyor (Pacman): %s\n", name)
		cmd = exec.Command("sudo", "pacman", "-Rns", "--noconfirm", name)
	}

	// Hata olsa bile (paket yoksa) devam etsin
	_ = cmd.Run()
	return nil
}

func (a *ArchLinuxProvider) Check(name string) (bool, error) {
	// Kontrol için her zaman pacman -Qi yeterlidir,
	// çünkü AUR paketleri de pacman veritabanına kaydolur.
	cmd := exec.Command("pacman", "-Qi", name)
	if err := cmd.Run(); err != nil {
		return false, nil // Paket yok
	}
	return true, nil // Paket var
}
