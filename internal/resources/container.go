package resources

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ContainerResource struct {
	Name  string
	Image string
	Ports []string
	State string // "running", "stopped", "absent"
}

func (c *ContainerResource) ID() string {
	return fmt.Sprintf("container:%s", c.Name)
}

func (c *ContainerResource) Check() (bool, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", c.Name)
	out, err := cmd.Output()

	isRunning := false
	if err == nil && strings.TrimSpace(string(out)) == "true" {
		isRunning = true
	}

	if c.State == "running" && isRunning {
		return true, nil
	}
	if c.State == "stopped" && !isRunning {
		return true, nil
	}
	if c.State == "absent" && err != nil {
		return true, nil
	}

	return false, nil
}

func (c *ContainerResource) Apply() error {
	if c.State == "absent" {
		exec.Command("docker", "rm", "-f", c.Name).Run()
		return nil
	}

	exec.Command("docker", "rm", "-f", c.Name).Run()

	args := []string{"run", "-d", "--name", c.Name}
	for _, p := range c.Ports {
		args = append(args, "-p", p)
	}
	args = append(args, c.Image)

	cmd := exec.Command("docker", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker run hatası: %s", string(out))
	}
	return nil
}

func (c *ContainerResource) Diff() (string, error) {
	return fmt.Sprintf("~ container: %s durumu %s yapılacak", c.Name, c.State), nil
}

func (c *ContainerResource) Undo(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return exec.Command("docker", "rm", "-f", c.Name).Run()
}
