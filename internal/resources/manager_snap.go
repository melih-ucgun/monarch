package resources

import "os/exec"

type SnapManager struct{}

func (m *SnapManager) IsInstalled(name string) (bool, error) {
	// snap list <name> -> Yüklüyse 0, değilse 1
	cmd := exec.Command("snap", "list", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *SnapManager) Install(name string) error {
	// Snap kurulumları genelde interaktif değildir ama garanti olsun.
	// Bazı uygulamalar "--classic" ister, şu anki yapıda bunu parametrik yapmak zor.
	// İleride config'e "options" eklenirse burası güncellenebilir.
	return exec.Command("snap", "install", name).Run()
}

func (m *SnapManager) Remove(name string) error {
	return exec.Command("snap", "remove", name).Run()
}
