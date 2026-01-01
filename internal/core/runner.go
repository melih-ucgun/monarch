package core

import (
	"os/exec"
)

// Runner interface defines methods for running commands.
// It allows mocking command execution in tests across all adapters.
type Runner interface {
	Run(cmd *exec.Cmd) error
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
	Output(cmd *exec.Cmd) ([]byte, error)
}

// RealRunner implements Runner using real os/exec.
type RealRunner struct{}

func (r *RealRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (r *RealRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func (r *RealRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

// CommandRunner is the global runner instance.
// Tests can replace this with a mock.
var CommandRunner Runner = &RealRunner{}

// RunCommand, bir komutu çalıştırır ve çıktısını/hatasını döner.
// Global Runner üzerinden çalışarak soyutlama sağlar.
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := CommandRunner.CombinedOutput(cmd)
	return string(out), err
}

// IsCommandAvailable, bir komutun sistemde yüklü olup olmadığını kontrol eder.
var IsCommandAvailable = func(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
