package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestFullFeatures runs an end-to-end test using the `full_features.yaml` configuration.
// It builds the application (or runs via go run) and asserts the side effects on the filesystem.
func TestFullFeatures(t *testing.T) {
	// 1. Setup Paths
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Assuming the test runs from the project root or tests/ directory
	// We need to locate the root of the project.
	// If running from tests/, parent is root. If running from root, wd is root.
	projectRoot := wd
	if strings.HasSuffix(wd, "tests") {
		projectRoot = filepath.Dir(wd)
	}

	configFile := filepath.Join(projectRoot, "tests", "full_features.yaml")
	vetoHome := filepath.Join(os.TempDir(), "veto_test_home")
	testTargetDir := "/tmp/veto_test_run" // Must match the variable in yaml

	// Clean up previous runs
	os.RemoveAll(vetoHome)
	os.RemoveAll(testTargetDir)
	defer func() {
		// Cleanup after test
		os.RemoveAll(vetoHome)
		os.RemoveAll(testTargetDir)
	}()

	// 2. Prepare Command
	// We use "go run ." to execute the current code without a separate build step
	// This ensures we test exactly what is in the codebase.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("go", "run", ".", "apply", configFile, "--no-snapshot")
	} else {
		cmd = exec.Command("go", "run", ".", "apply", configFile, "--no-snapshot")
	}

	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "VETO_HOME="+vetoHome)

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Veto apply failed: %v\nOutput:\n%s", err, string(output))
	} else {
		t.Logf("Veto apply output:\n%s", string(output))
	}

	// 3. Verifications

	// Verify File Existence and Content
	targetFile := filepath.Join(testTargetDir, "config.txt")
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Errorf("Failed to read created file: %v", err)
	} else {
		expected := "Hello Veto CI"
		if !strings.Contains(string(content), expected) {
			t.Errorf("File content mismatch. Expected to contain %q, got:\n%s", expected, string(content))
		}
	}

	// Verify Symlink
	linkFile := filepath.Join(testTargetDir, "config_link.txt")
	info, err := os.Lstat(linkFile)
	if err != nil {
		t.Errorf("Symlink missing: %v", err)
	} else {
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("Expected symlink, got regular file")
		}
	}

	// Verify Git Clone
	gitDir := filepath.Join(testTargetDir, "veto-repo", ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("Git repo was not cloned successfully")
	}
}
