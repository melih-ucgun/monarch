package resources

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ContainerResource struct {
	CanonicalID string   `mapstructure:"-"`
	Name        string   `mapstructure:"name"`
	Image       string   `mapstructure:"image"`
	State       string   `mapstructure:"state"` // running, stopped, absent
	Ports       []string `mapstructure:"ports"` // "8080:80"
}

func (r *ContainerResource) ID() string {
	return r.CanonicalID
}

func (r *ContainerResource) Check() (bool, error) {
	// Docker inspect ile çalışıyor mu kontrol et
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", r.Name)
	out, err := cmd.Output()

	exists := (err == nil)
	isRunning := false
	if exists && strings.TrimSpace(string(out)) == "true" {
		isRunning = true
	}

	if r.State == "running" {
		return isRunning, nil
	} else if r.State == "stopped" {
		// Varsa ama durmuşsa true, çalışıyorsa false
		return exists && !isRunning, nil
	} else if r.State == "absent" {
		return !exists, nil
	}

	return false, fmt.Errorf("bilinmeyen container state: %s", r.State)
}

func (r *ContainerResource) Apply() error {
	// Önce container varsa temizle (Basit yaklaşım: Recreate)
	// TODO: Daha zeki bir update mekanizması (sadece image değiştiyse vs.) eklenebilir.
	existsCmd := exec.Command("docker", "inspect", r.Name)
	if err := existsCmd.Run(); err == nil {
		// Container var, kaldır
		rmCmd := exec.Command("docker", "rm", "-f", r.Name)
		if out, err := rmCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("eski container silinemedi: %s", string(out))
		}
	}

	if r.State == "absent" {
		return nil
	}

	if r.State == "stopped" {
		// Create but not start
		args := []string{"create", "--name", r.Name}
		for _, p := range r.Ports {
			args = append(args, "-p", p)
		}
		args = append(args, r.Image)
		if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("docker create hatası: %s", string(out))
		}
		return nil
	}

	// State: running
	args := []string{"run", "-d", "--name", r.Name}
	for _, p := range r.Ports {
		args = append(args, "-p", p)
	}
	args = append(args, r.Image)

	if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("docker run hatası: %s", string(out))
	}

	return nil
}

func (r *ContainerResource) Undo(ctx context.Context) error {
	return exec.Command("docker", "rm", "-f", r.Name).Run()
}

func (r *ContainerResource) Diff() (string, error) {
	return fmt.Sprintf("Container[%s] state mismatch", r.Name), nil
}
