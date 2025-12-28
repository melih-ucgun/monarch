package resources

import (
	"fmt"
	"os/exec"
)

type ExecResource struct {
	CanonicalID string
	Name        string
	Command     string
}

func (e *ExecResource) ID() string {
	return e.CanonicalID
}

func (e *ExecResource) Check() (bool, error) {
	return false, nil
}

func (e *ExecResource) Apply() error {
	cmd := exec.Command("bash", "-c", e.Command)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("komut çalıştırma hatası [%s]: %w", e.Name, err)
	}
	return nil
}
