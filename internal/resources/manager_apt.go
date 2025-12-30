package resources

import "os/exec"

type AptManager struct{}

func (m *AptManager) IsInstalled(name string) (bool, error) {
	// dpkg -s <paket> -> Yüklüyse 0, değilse 1 döner
	cmd := exec.Command("dpkg", "-s", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *AptManager) Install(name string) error {
	cmd := exec.Command("apt-get", "install", "-y", name)
	cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	return cmd.Run()
}

func (m *AptManager) Remove(name string) error {
	cmd := exec.Command("apt-get", "remove", "-y", name)
	cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	return cmd.Run()
}
