package resources

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ServiceResource struct {
	CanonicalID string
	Name        string
	State       string // "started", "stopped", "restarted"
	Enabled     bool   // systemctl enable
}

func (s *ServiceResource) ID() string {
	if s.CanonicalID != "" {
		return s.CanonicalID
	}
	return fmt.Sprintf("service:%s", s.Name)
}

func (s *ServiceResource) Check() (bool, error) {
	// 1. Aktiflik Kontrolü
	cmdActive := exec.Command("systemctl", "is-active", s.Name)
	errActive := cmdActive.Run()
	isActive := (errActive == nil)

	shouldBeActive := (s.State == "started" || s.State == "restarted")

	if isActive != shouldBeActive {
		return false, nil
	}

	// 2. Enable Kontrolü
	cmdEnabled := exec.Command("systemctl", "is-enabled", s.Name)
	output, _ := cmdEnabled.Output()
	isEnabled := strings.TrimSpace(string(output)) == "enabled"

	if isEnabled != s.Enabled {
		return false, nil
	}

	return true, nil
}

func (s *ServiceResource) Apply() error {
	// Enable/Disable
	if s.Enabled {
		exec.Command("sudo", "systemctl", "enable", s.Name).Run()
	} else {
		exec.Command("sudo", "systemctl", "disable", s.Name).Run()
	}

	// State
	action := ""
	switch s.State {
	case "started":
		action = "start"
	case "stopped":
		action = "stop"
	case "restarted":
		action = "restart"
	}

	if action != "" {
		out, err := exec.Command("sudo", "systemctl", action, s.Name).CombinedOutput()
		if err != nil {
			return fmt.Errorf("servis işlemi (%s) hatası: %s", action, string(out))
		}
	}
	return nil
}

func (s *ServiceResource) Diff() (string, error) {
	return fmt.Sprintf("~ service: %s -> %s, Enabled: %v", s.Name, s.State, s.Enabled), nil
}

func (s *ServiceResource) Undo(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Undo: Servisi durdur ve disable et
	return exec.Command("sudo", "systemctl", "disable", "--now", s.Name).Run()
}
