package service

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/melih-ucgun/monarch/internal/core"
)

type ServiceAdapter struct {
	core.BaseResource
	State   string // active, stopped, restarted
	Enabled bool   // true, false
}

func NewServiceAdapter(name string, params map[string]interface{}) *ServiceAdapter {
	state, _ := params["state"].(string)
	if state == "" {
		state = "active"
	}

	enabled := true
	if e, ok := params["enabled"].(bool); ok {
		enabled = e
	}

	return &ServiceAdapter{
		BaseResource: core.BaseResource{Name: name, Type: "service"},
		State:        state,
		Enabled:      enabled,
	}
}

func (r *ServiceAdapter) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("service name is required")
	}
	return nil
}

func (r *ServiceAdapter) Check(ctx *core.SystemContext) (bool, error) {
	// 1. Enable Durumu Kontrolü
	cmdEnable := exec.Command("systemctl", "is-enabled", r.Name)
	outEnable, _ := cmdEnable.CombinedOutput()
	isEnabled := strings.TrimSpace(string(outEnable)) == "enabled"

	if r.Enabled != isEnabled {
		return true, nil
	}

	// 2. Active Durumu Kontrolü
	cmdActive := exec.Command("systemctl", "is-active", r.Name)
	outActive, _ := cmdActive.CombinedOutput()
	isActive := strings.TrimSpace(string(outActive)) == "active"

	shouldBeActive := (r.State == "active" || r.State == "started")

	if isActive != shouldBeActive {
		return true, nil
	}

	return false, nil
}

func (r *ServiceAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Service %s is in desired state", r.Name)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] Configure service %s (enable=%v, state=%s)", r.Name, r.Enabled, r.State)), nil
	}

	messages := []string{}

	// Enable/Disable
	actionEnable := "disable"
	if r.Enabled {
		actionEnable = "enable"
	}

	// Sadece idempotent olması için force uygulayabiliriz veya check sonucuna göre yapabiliriz.
	// systemctl enable/disable genellikle idempotenttir.
	if out, err := exec.Command("systemctl", actionEnable, "--now", r.Name).CombinedOutput(); err != nil {
		return core.Failure(err, fmt.Sprintf("Failed to %s service: %s", actionEnable, string(out))), err
	}
	messages = append(messages, fmt.Sprintf("Service %sd", actionEnable))

	// Start/Stop
	actionState := ""
	if r.State == "active" || r.State == "started" {
		actionState = "start"
	} else if r.State == "stopped" {
		actionState = "stop"
	} else if r.State == "restarted" {
		actionState = "restart"
	}

	if actionState != "" {
		if out, err := exec.Command("systemctl", actionState, r.Name).CombinedOutput(); err != nil {
			return core.Failure(err, fmt.Sprintf("Failed to %s service: %s", actionState, string(out))), err
		}
		messages = append(messages, fmt.Sprintf("Service %sed", actionState))
	}

	return core.SuccessChange(strings.Join(messages, ", ")), nil
}
