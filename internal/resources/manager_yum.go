package resources

import "os/exec"

type YumManager struct{}

func (m *YumManager) IsInstalled(name string) (bool, error) {
	cmd := exec.Command("rpm", "-q", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *YumManager) Install(name string) error {
	return exec.Command("yum", "install", "-y", name).Run()
}

func (m *YumManager) Remove(name string) error {
	return exec.Command("yum", "remove", "-y", name).Run()
}
