package pkg

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/melih-ucgun/veto/internal/core"
)

// MockRunner is a mock implementation of core.Runner interface.
type MockRunner struct {
	RunFunc            func(cmd *exec.Cmd) error
	CombinedOutputFunc func(cmd *exec.Cmd) ([]byte, error)
	OutputFunc         func(cmd *exec.Cmd) ([]byte, error)
}

func (m *MockRunner) Run(cmd *exec.Cmd) error {
	if m.RunFunc != nil {
		return m.RunFunc(cmd)
	}
	return nil
}

func (m *MockRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	if m.CombinedOutputFunc != nil {
		return m.CombinedOutputFunc(cmd)
	}
	return []byte{}, nil
}

func (m *MockRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	if m.OutputFunc != nil {
		return m.OutputFunc(cmd)
	}
	return []byte{}, nil
}

func TestPacmanAdapter_Check(t *testing.T) {
	// Restore original runner after tests
	defer func() { core.CommandRunner = &core.RealRunner{} }()

	tests := []struct {
		name          string
		packageName   string
		state         string
		mockRunErr    error
		expectedCheck bool
	}{
		{
			name:          "Package not installed, State=present -> Needs Action (Types.True)",
			packageName:   "git",
			state:         "present",
			mockRunErr:    errors.New("not found"), // simule "pacman -Qi" failing
			expectedCheck: true,
		},
		{
			name:          "Package installed, State=present -> No Action (Types.False)",
			packageName:   "git",
			state:         "present",
			mockRunErr:    nil, // simule "pacman -Qi" success
			expectedCheck: false,
		},
		{
			name:          "Package installed, State=absent -> Needs Action (Types.True)",
			packageName:   "vim",
			state:         "absent",
			mockRunErr:    nil,
			expectedCheck: true,
		},
		{
			name:          "Package not installed, State=absent -> No Action (Types.False)",
			packageName:   "vim",
			state:         "absent",
			mockRunErr:    errors.New("not found"),
			expectedCheck: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Mock
			core.CommandRunner = &MockRunner{
				RunFunc: func(cmd *exec.Cmd) error {
					// Verify command is checking package existence
					if len(cmd.Args) < 1 || cmd.Args[0] != "pacman" {
						return fmt.Errorf("unexpected command: %v", cmd.Args)
					}
					// Only check -Qi if it's a check command
					if len(cmd.Args) >= 2 && cmd.Args[1] == "-Qi" {
						return tt.mockRunErr
					}
					return fmt.Errorf("unexpected args: %v", cmd.Args)
				},
			}

			adapter := NewPacmanAdapter(tt.packageName, map[string]interface{}{"state": tt.state}).(*PacmanAdapter)
			needsAction, err := adapter.Check(&core.SystemContext{})

			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if needsAction != tt.expectedCheck {
				t.Errorf("Check() = %v, want %v", needsAction, tt.expectedCheck)
			}
		})
	}
}

func TestPacmanAdapter_Apply(t *testing.T) {
	defer func() { core.CommandRunner = &core.RealRunner{} }()

	t.Run("DryRun should not execute install command", func(t *testing.T) {
		adapter := NewPacmanAdapter("htop", map[string]interface{}{"state": "present"}).(*PacmanAdapter)

		// Mock check to say package is missing (so it tries to install)
		core.CommandRunner = &MockRunner{
			RunFunc: func(cmd *exec.Cmd) error {
				return errors.New("not installed")
			},
		}

		ctx := &core.SystemContext{DryRun: true}
		result, err := adapter.Apply(ctx)

		if err != nil {
			t.Fatalf("Apply returned error: %v", err)
		}
		if !result.Changed {
			t.Errorf("Expected Changed=true for DryRun")
		}
		if !strings.Contains(result.Message, "DryRun") {
			t.Errorf("Expected DryRun message, got: %s", result.Message)
		}
	})

	t.Run("Install success", func(t *testing.T) {
		adapter := NewPacmanAdapter("htop", map[string]interface{}{"state": "present"}).(*PacmanAdapter)

		var executedCmd []string

		core.CommandRunner = &MockRunner{
			RunFunc: func(cmd *exec.Cmd) error {
				// 1. Check call (returns err -> not installed)
				if cmd.Args[1] == "-Qi" {
					return errors.New("not installed")
				}
				return nil
			},
			CombinedOutputFunc: func(cmd *exec.Cmd) ([]byte, error) {
				// 2. Install call
				executedCmd = cmd.Args
				return []byte("installation success"), nil
			},
		}

		ctx := &core.SystemContext{DryRun: false}
		result, err := adapter.Apply(ctx)

		if err != nil {
			t.Fatalf("Apply returned error: %v", err)
		}
		if !result.Changed {
			t.Error("Expected Changed=true")
		}

		// Verify command args for install: pacman -S --noconfirm --needed htop
		expected := []string{"pacman", "-S", "--noconfirm", "--needed", "htop"}
		if len(executedCmd) != 5 {
			t.Errorf("Expected command length 5, got %d: %v", len(executedCmd), executedCmd)
		}
		if executedCmd[0] != expected[0] || executedCmd[4] != expected[4] {
			t.Errorf("Unexpected command: %v", executedCmd)
		}
	})
}

func TestPacmanAdapter_Revert(t *testing.T) {
	defer func() { core.CommandRunner = &core.RealRunner{} }()

	t.Run("Revert installed package", func(t *testing.T) {
		adapter := NewPacmanAdapter("nano", map[string]interface{}{"state": "present"}).(*PacmanAdapter)
		adapter.ActionPerformed = "installed"

		var executedCmd []string

		core.CommandRunner = &MockRunner{
			CombinedOutputFunc: func(cmd *exec.Cmd) ([]byte, error) {
				executedCmd = cmd.Args
				return []byte("removed"), nil
			},
		}

		err := adapter.Revert(&core.SystemContext{})
		if err != nil {
			t.Fatalf("Revert failed: %v", err)
		}

		// Verify remove command: pacman -Rns --noconfirm nano
		expected := []string{"pacman", "-Rns", "--noconfirm", "nano"}
		if len(executedCmd) != 4 {
			t.Errorf("Expected command length 4, got %d: %v", len(executedCmd), executedCmd)
		}
		if executedCmd[1] != expected[1] || executedCmd[3] != expected[3] {
			t.Errorf("Unexpected command: got %v, want %v", executedCmd, expected)
		}
	})
}
