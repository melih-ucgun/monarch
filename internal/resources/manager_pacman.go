package resources

import "os/exec"

type PacmanManager struct{}

func (m *PacmanManager) IsInstalled(name string) (bool, error) {
	// pacman -Q <paket> -> Yüklüyse 0, değilse 1 döner
	cmd := exec.Command("pacman", "-Q", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *PacmanManager) Install(name string) error {
	// --noconfirm: Onay istemeden kur, --needed: Zaten güncelse atla
	return exec.Command("pacman", "-S", "--noconfirm", "--needed", name).Run()
}

func (m *PacmanManager) Remove(name string) error {
	return exec.Command("pacman", "-Rns", "--noconfirm", name).Run()
}
