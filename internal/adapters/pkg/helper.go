package pkg

import (
	"os/exec"
)

// isInstalled, verilen komutun başarıyla çalışıp çalışmadığını kontrol eder.
// Paket yöneticileri genellikle paket varsa 0, yoksa hata kodu döner.
func isInstalled(checkCmd string, args ...string) bool {
	cmd := exec.Command(checkCmd, args...)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// runCommand, bir komutu çalıştırır ve çıktısını/hatasını döner.
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
