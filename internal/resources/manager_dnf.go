package resources

import "os/exec"

type DnfManager struct{}

func (m *DnfManager) IsInstalled(name string) (bool, error) {
	// RPM tabanlı sistemlerde 'rpm -q' en hızlı ve tutarlı yöntemdir.
	// DNF veritabanını sorgulamak yavaş olabilir.
	cmd := exec.Command("rpm", "-q", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *DnfManager) Install(name string) error {
	return exec.Command("dnf", "install", "-y", name).Run()
}

func (m *DnfManager) Remove(name string) error {
	return exec.Command("dnf", "remove", "-y", name).Run()
}
