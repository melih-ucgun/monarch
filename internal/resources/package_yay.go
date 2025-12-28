package resources

import "os/exec"

type YayProvider struct{}

func (p *YayProvider) IsInstalled(name string) (bool, error) {
	// pacman -Q paketin kurulu olup olmadığını kontrol eder
	err := exec.Command("yay", "-Q", name).Run()
	return err == nil, nil
}

func (p *YayProvider) Install(name string) error {
	// --noconfirm ile onay sormadan kurar
	return exec.Command("yay", "-S", "--noconfirm", name).Run()
}

func (p *YayProvider) Remove(name string) error {
	return exec.Command("yay", "-R", "--noconfirm", name).Run()
}
