package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProfileManager struct {
	BaseDir     string
	ProfilesDir string
	ActiveFile  string
}

func NewProfileManager(baseDir string) *ProfileManager {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".monarch")
	}
	return &ProfileManager{
		BaseDir:     baseDir,
		ProfilesDir: filepath.Join(baseDir, "profiles"),
		ActiveFile:  filepath.Join(baseDir, "active_profile"),
	}
}

// EnsureDirs creates base directories
func (m *ProfileManager) EnsureDirs() error {
	return os.MkdirAll(m.ProfilesDir, 0755)
}

// Create creates a new profile directory and a default main.yaml
func (m *ProfileManager) Create(name string) error {
	if err := m.EnsureDirs(); err != nil {
		return err
	}

	profileDir := filepath.Join(m.ProfilesDir, name)
	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return err
	}

	defaultConfig := fmt.Sprintf("# Profile: %s\nresources: []\n", name)
	configFile := filepath.Join(profileDir, "main.yaml")

	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}

// List returns all profile names
func (m *ProfileManager) List() ([]string, error) {
	if err := m.EnsureDirs(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(m.ProfilesDir)
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, e := range entries {
		if e.IsDir() {
			profiles = append(profiles, e.Name())
		}
	}
	return profiles, nil
}

// Use sets the active profile
func (m *ProfileManager) Use(name string) error {
	// Verify it exists
	profileDir := filepath.Join(m.ProfilesDir, name)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	return os.WriteFile(m.ActiveFile, []byte(name), 0644)
}

// GetActive returns the name of the active profile
func (m *ProfileManager) GetActive() (string, error) {
	content, err := os.ReadFile(m.ActiveFile)
	if os.IsNotExist(err) {
		return "", nil // No active profile
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// GetProfilePath returns the path to main.yaml of the given profile (or active if empty)
func (m *ProfileManager) GetProfilePath(name string) (string, error) {
	if name == "" {
		active, err := m.GetActive()
		if err != nil {
			return "", err
		}
		if active == "" {
			return "", nil // No active profile
		}
		name = active
	}
	return filepath.Join(m.ProfilesDir, name, "main.yaml"), nil
}
