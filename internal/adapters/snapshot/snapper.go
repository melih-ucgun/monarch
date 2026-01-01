package snapshot

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
)

// Snapper manages integration with the snapper tool
type Snapper struct {
	configName string
}

func NewSnapper() *Snapper {
	return &Snapper{
		configName: "root",
	}
}

func (s *Snapper) Name() string {
	return "Snapper"
}

func (s *Snapper) IsAvailable() bool {
	_, err := exec.LookPath("snapper")
	return err == nil
}

func (s *Snapper) CreateSnapshot(description string) error {
	cmd := exec.Command("snapper", "-c", s.configName, "create", "-d", description)
	return cmd.Run()
}

func (s *Snapper) CreatePreSnapshot(description string) (string, error) {
	cmd := exec.Command("snapper", "-c", s.configName, "create", "-t", "pre", "-p", "-d", description)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("snapper pre failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *Snapper) CreatePostSnapshot(id string, description string) error {
	if id == "" {
		return fmt.Errorf("invalid pre-snapshot id")
	}
	cmd := exec.Command("snapper", "-c", s.configName, "create", "-t", "post", "--pre-number", id, "-d", description)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("snapper post failed: %w", err)
	}
	pterm.Success.Printf("Snapper pair created (Pre: %s)\n", id)
	return nil
}
