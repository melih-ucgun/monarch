package resources

import "os/exec"

type ParuManager struct{}

func (m *ParuManager) IsInstalled(name string) (bool, error) {
	// paru -Q da pacman -Q gibi çalışır
	cmd := exec.Command("paru", "-Q", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *ParuManager) Install(name string) error {
	// --noconfirm: Onaylama
	// --needed: Güncelse atla
	return exec.Command("paru", "-S", "--noconfirm", "--needed", name).Run()
}

func (m *ParuManager) Remove(name string) error {
	// -Rns: Recursive, no-save
	return exec.Command("paru", "-Rns", "--noconfirm", name).Run()
}
