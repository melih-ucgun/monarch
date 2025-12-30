package resources

import "os/exec"

type ApkManager struct{}

func (m *ApkManager) IsInstalled(name string) (bool, error) {
	// apk -e info <paket> -> Yüklüyse 0, değilse 1 exit code döner
	cmd := exec.Command("apk", "-e", "info", name)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *ApkManager) Install(name string) error {
	// --no-cache: Paketi kur ama index dosyasını diske kaydetme (image boyutunu küçük tutar)
	// Normal kurulum için sadece "add" de kullanılabilir ama best-practice budur.
	return exec.Command("apk", "add", "--no-cache", name).Run()
}

func (m *ApkManager) Remove(name string) error {
	return exec.Command("apk", "del", name).Run()
}
