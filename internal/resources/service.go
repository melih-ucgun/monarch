package resources

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ServiceResource struct {
	CanonicalID string `mapstructure:"-"`
	Name        string `mapstructure:"name"`
	State       string `mapstructure:"state"`   // started, stopped, restarted
	Enabled     bool   `mapstructure:"enabled"` // enable on boot
}

func (r *ServiceResource) ID() string {
	return r.CanonicalID
}

func (r *ServiceResource) Check() (bool, error) {
	// 1. Active Durumu Kontrolü
	// systemctl is-active -> active (exit 0), inactive (exit 3)
	cmdActive := exec.Command("systemctl", "is-active", r.Name)
	errActive := cmdActive.Run()

	isActive := (errActive == nil)
	shouldBeActive := (r.State == "started" || r.State == "restarted")

	if isActive != shouldBeActive {
		return false, nil
	}

	// 2. Enabled Durumu Kontrolü
	// systemctl is-enabled -> enabled (exit 0), disabled (exit 1)
	cmdEnabled := exec.Command("systemctl", "is-enabled", r.Name)
	outEnabled, _ := cmdEnabled.CombinedOutput()
	isEnabled := (strings.TrimSpace(string(outEnabled)) == "enabled")

	if isEnabled != r.Enabled {
		return false, nil
	}

	return true, nil
}

func (r *ServiceResource) Apply() error {
	// State değişikliği
	var action string
	switch r.State {
	case "started":
		action = "start"
	case "stopped":
		action = "stop"
	case "restarted":
		action = "restart"
	}

	if action != "" {
		if out, err := exec.Command("systemctl", action, r.Name).CombinedOutput(); err != nil {
			return fmt.Errorf("systemctl %s hatası: %s", action, string(out))
		}
	}

	// Enable/Disable değişikliği
	enableAction := "disable"
	if r.Enabled {
		enableAction = "enable"
	}

	// --now flag'i hem enable edip hem başlatabilir ama biz state'i ayrı yönettik
	if out, err := exec.Command("systemctl", enableAction, r.Name).CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl %s hatası: %s", enableAction, string(out))
	}

	return nil
}

func (r *ServiceResource) Undo(ctx context.Context) error {
	// Undo için basitçe durduruyoruz
	exec.Command("systemctl", "stop", r.Name).Run()
	return nil
}

func (r *ServiceResource) Diff() (string, error) {
	return fmt.Sprintf("Service[%s] state/enabled mismatch", r.Name), nil
}
