package resources

import (
	"fmt"
	"os/exec"
	"strings"
)

type ServiceResource struct {
	CanonicalID  string
	ServiceName  string
	DesiredState string
	Enabled      bool
}

func (s *ServiceResource) ID() string {
	return s.CanonicalID
}

func (s *ServiceResource) Check() (bool, error) {
	isActiveCmd := exec.Command("systemctl", "is-active", s.ServiceName)
	err := isActiveCmd.Run()
	actualState := "stopped"
	if err == nil {
		actualState = "running"
	}

	isEnabledCmd := exec.Command("systemctl", "is-enabled", s.ServiceName)
	enabledOut, _ := isEnabledCmd.Output()
	actualEnabled := strings.TrimSpace(string(enabledOut)) == "enabled"

	return (actualState == s.DesiredState) && (actualEnabled == s.Enabled), nil
}

func (s *ServiceResource) Apply() error {
	action := "start"
	if s.DesiredState == "stopped" {
		action = "stop"
	}

	if err := exec.Command("sudo", "systemctl", action, s.ServiceName).Run(); err != nil {
		return fmt.Errorf("servis %s yapılamadı: %w", action, err)
	}

	enableAction := "enable"
	if !s.Enabled {
		enableAction = "disable"
	}

	if err := exec.Command("sudo", "systemctl", enableAction, s.ServiceName).Run(); err != nil {
		return fmt.Errorf("servis %s yapılamadı: %w", enableAction, err)
	}

	return nil
}
