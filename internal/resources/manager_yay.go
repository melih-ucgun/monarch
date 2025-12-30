package resources

import "os/exec"

type YayManager struct{}

func (m *YayManager) IsInstalled(name string) (bool, error) {
	cmd := exec.Command("yay", "-Q", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *YayManager) Install(name string) error {
	// --noconfirm: Onaylama
	// --needed: Güncelse atla
	return exec.Command("yay", "-S", "--noconfirm", "--needed", name).Run()
}

func (m *YayManager) Remove(name string) error {
	return exec.Command("yay", "-Rns", "--noconfirm", name).Run()
}
