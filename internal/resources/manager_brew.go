package resources

import "os/exec"

type BrewManager struct{}

func (m *BrewManager) IsInstalled(name string) (bool, error) {
	cmd := exec.Command("brew", "list", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *BrewManager) Install(name string) error {
	return exec.Command("brew", "install", name).Run()
}

func (m *BrewManager) Remove(name string) error {
	return exec.Command("brew", "uninstall", name).Run()
}
