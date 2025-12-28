package resources

import (
	"fmt"
	"os/exec"
	"strings"
)

type ServiceResource struct {
	ServiceName  string
	DesiredState string // "running" veya "stopped"
	Enabled      bool
}

func (s *ServiceResource) ID() string {
	return fmt.Sprintf("svc:%s", s.ServiceName)
}

func (s *ServiceResource) Check() (bool, error) {
	// 1. Aktiflik Kontrolü (is-active)
	isActiveCmd := exec.Command("systemctl", "is-active", s.ServiceName)
	err := isActiveCmd.Run()
	actualState := "stopped"
	if err == nil {
		actualState = "running"
	}

	// 2. Başlangıç Durumu Kontrolü (is-enabled)
	isEnabledCmd := exec.Command("systemctl", "is-enabled", s.ServiceName)
	enabledOut, _ := isEnabledCmd.Output()
	actualEnabled := strings.TrimSpace(string(enabledOut)) == "enabled"

	// Hem durum hem de enable bilgisi eşleşmeli
	return (actualState == s.DesiredState) && (actualEnabled == s.Enabled), nil
}

func (s *ServiceResource) Apply() error {
	// Durumu ayarla (start/stop)
	action := "start"
	if s.DesiredState == "stopped" {
		action = "stop"
	}

	if err := exec.Command("sudo", "systemctl", action, s.ServiceName).Run(); err != nil {
		return fmt.Errorf("servis %s yapılamadı: %w", action, err)
	}

	// Başlangıç ayarını yap (enable/disable)
	enableAction := "enable"
	if !s.Enabled {
		enableAction = "disable"
	}

	if err := exec.Command("sudo", "systemctl", enableAction, s.ServiceName).Run(); err != nil {
		return fmt.Errorf("servis %s yapılamadı: %w", enableAction, err)
	}

	return nil
}
