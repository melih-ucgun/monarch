package resources

import (
	"fmt"
	"os/exec"
	"strings"
)

// ContainerResource, Docker veya Podman konteynerlerini yönetir.
type ContainerResource struct {
	CanonicalID string
	Name        string
	Image       string
	State       string   // running, stopped, absent
	Ports       []string // ["8080:80"]
	Env         []string // ["KEY=VAL"]
	Volumes     []string // ["/host:/cont"]
	Engine      string   // docker veya podman
}

func (c *ContainerResource) ID() string {
	return c.CanonicalID
}

// Check, konteynerin mevcut durumunu kontrol eder.
func (c *ContainerResource) Check() (bool, error) {
	cmd := exec.Command(c.Engine, "inspect", "--format", "{{.State.Running}}", c.Name)
	output, err := cmd.CombinedOutput()

	exists := err == nil
	isRunning := strings.TrimSpace(string(output)) == "true"

	switch c.State {
	case "absent":
		return !exists, nil
	case "stopped":
		return exists && !isRunning, nil
	case "running":
		return exists && isRunning, nil
	default:
		return exists && isRunning, nil
	}
}

// Diff, beklenen ve mevcut durum arasındaki farkı metin olarak döner.
func (c *ContainerResource) Diff() (string, error) {
	inState, _ := c.Check()
	if inState {
		return "", nil
	}
	return fmt.Sprintf("! container: %s (İstenen: %s, İmaj: %s)", c.Name, c.State, c.Image), nil
}

// Apply, konteyneri istenen duruma getirir (silme, durdurma veya yeniden başlatma).
func (c *ContainerResource) Apply() error {
	// Temiz bir kurulum için mevcut olanı sil (Declarative yaklaşım)
	exec.Command(c.Engine, "rm", "-f", c.Name).Run()

	if c.State == "absent" {
		return nil
	}

	args := []string{"run", "-d", "--name", c.Name}
	for _, p := range c.Ports {
		args = append(args, "-p", p)
	}
	for _, e := range c.Env {
		args = append(args, "-e", e)
	}
	for _, v := range c.Volumes {
		args = append(args, "-v", v)
	}
	args = append(args, c.Image)

	if err := exec.Command(c.Engine, args...).Run(); err != nil {
		return fmt.Errorf("konteyner başlatılamadı: %w", err)
	}

	if c.State == "stopped" {
		return exec.Command(c.Engine, "stop", c.Name).Run()
	}

	return nil
}

// GetContainerEngine, sistemdeki konteyner motorunu tespit eder.
func GetContainerEngine() string {
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman"
	}
	return "docker"
}
