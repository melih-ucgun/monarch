package resources

import "os/exec"

type ParuProvider struct{}

func (p *ParuProvider) IsInstalled(name string) (bool, error) {
	// pacman -Q paketin kurulu olup olmadığını kontrol eder
	err := exec.Command("paru", "-Q", name).Run()
	return err == nil, nil
}

func (p *ParuProvider) Install(name string) error {
	// --noconfirm ile onay sormadan kurar
	return exec.Command("paru", "-S", "--noconfirm", name).Run()
}

func (p *ParuProvider) Remove(name string) error {
	return exec.Command("paru", "-R", "--noconfirm", name).Run()
}
