package resources

import (
	"fmt"
	"os"
	"os/exec"
)

type ExecResource struct {
	CanonicalID string
	Name        string
	Command     string
	Creates     string
	OnlyIf      string
	Unless      string
}

func (e *ExecResource) ID() string {
	return e.CanonicalID
}

func (e *ExecResource) Check() (bool, error) {
	if e.Creates != "" {
		if _, err := os.Stat(e.Creates); err == nil {
			return true, nil
		}
	}
	if e.OnlyIf != "" {
		if err := exec.Command("bash", "-c", e.OnlyIf).Run(); err != nil {
			return true, nil
		}
	}
	if e.Unless != "" {
		if err := exec.Command("bash", "-c", e.Unless).Run(); err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (e *ExecResource) Diff() (string, error) {
	skip, _ := e.Check()
	if skip {
		return "", nil
	}
	return fmt.Sprintf("! exec: %s (Çalıştırılacak)", e.Name), nil
}

func (e *ExecResource) Apply() error {
	cmd := exec.Command("bash", "-c", e.Command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
