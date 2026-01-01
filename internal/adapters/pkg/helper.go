package pkg

import (
	"fmt"
	"os/exec"

	"github.com/melih-ucgun/veto/internal/core"
)

func init() {
	core.RegisterResource("package", DetectPackageManager)
	core.RegisterResource("pkg", DetectPackageManager)
}

// DetectPackageManager detects and returns the appropriate package manager adapter.
func DetectPackageManager(name string, params map[string]interface{}, ctx *core.SystemContext) (core.Resource, error) {
	state, _ := params["state"].(string)
	if state == "" {
		state = "present"
	}

	switch ctx.Distro {
	case "arch", "cachyos", "manjaro", "endeavouros":
		return NewPacmanAdapter(name, params), nil
	case "ubuntu", "debian", "pop", "mint", "kali":
		return NewAptAdapter(name, params), nil
	case "fedora", "rhel", "centos", "almalinux":
		return NewDnfAdapter(name, params), nil
	case "alpine":
		return NewApkAdapter(name, params), nil
	case "opensuse", "sles":
		return NewZypperAdapter(name, params), nil
	case "darwin":
		return NewBrewAdapter(name, params), nil
	default:
		// Fallback to searching available commands
		if core.IsCommandAvailable("pacman") {
			return NewPacmanAdapter(name, params), nil
		} else if core.IsCommandAvailable("apt-get") {
			return NewAptAdapter(name, params), nil
		} else if core.IsCommandAvailable("dnf") {
			return NewDnfAdapter(name, params), nil
		}

		return nil, fmt.Errorf("automatic package manager detection failed for distro: %s", ctx.Distro)
	}
}

// isInstalled, verilen komutun başarıyla çalışıp çalışmadığını kontrol eder.
// Paket yöneticileri genellikle paket varsa 0, yoksa hata kodu döner.
func isInstalled(checkCmd string, args ...string) bool {
	cmd := exec.Command(checkCmd, args...)
	if err := core.CommandRunner.Run(cmd); err != nil {
		return false
	}
	return true
}

// runCommand, bir komutu çalıştırır ve çıktısını/hatasını döner.
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := core.CommandRunner.CombinedOutput(cmd)
	return string(out), err
}
