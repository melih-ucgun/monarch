package resources

import "os/exec"

type ZypperManager struct{}

func (m *ZypperManager) IsInstalled(name string) (bool, error) {
	// openSUSE de RPM tabanlıdır, bu yüzden rpm -q güvenilirdir.
	cmd := exec.Command("rpm", "-q", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *ZypperManager) Install(name string) error {
	// -n: Non-interactive (soru sorma, varsayılanı kabul et)
	return exec.Command("zypper", "-n", "install", name).Run()
}

func (m *ZypperManager) Remove(name string) error {
	return exec.Command("zypper", "-n", "remove", name).Run()
}
