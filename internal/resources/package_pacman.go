package resources

import "os/exec"

type PacmanProvider struct{}

func (p *PacmanProvider) IsInstalled(name string) (bool, error) {
	// pacman -Q paketin kurulu olup olmadığını kontrol eder
	err := exec.Command("pacman", "-Q", name).Run()
	return err == nil, nil
}

func (p *PacmanProvider) Install(name string) error {
	// --noconfirm ile onay sormadan kurar
	return exec.Command("sudo", "pacman", "-S", "--noconfirm", name).Run()
}

func (p *PacmanProvider) Remove(name string) error {
	return exec.Command("sudo", "pacman", "-R", "--noconfirm", name).Run()
}
