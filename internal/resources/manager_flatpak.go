package resources

import "os/exec"

type FlatpakManager struct{}

func (m *FlatpakManager) IsInstalled(name string) (bool, error) {
	// flatpak info <app-id> -> Yüklüyse 0, değilse 1 döner
	cmd := exec.Command("flatpak", "info", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *FlatpakManager) Install(name string) error {
	// -y: Onaylama (non-interactive)
	// --noninteractive: Ekstra güvenlik, soru sormaz
	// Not: Genelde "flatpak install flathub <name>" kullanılır ama
	// remote belirtmeden de arayıp kurabilir.
	return exec.Command("flatpak", "install", "-y", "--noninteractive", name).Run()
}

func (m *FlatpakManager) Remove(name string) error {
	return exec.Command("flatpak", "uninstall", "-y", name).Run()
}
