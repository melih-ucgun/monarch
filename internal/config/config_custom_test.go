package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestResourceConfig_UnmarshalYAML_Naming(t *testing.T) {
	tests := []struct {
		name           string
		yamlData       string
		envDistro      string
		expectedName   string
		expectedID     string // checking if other fields are preserved
	}{
		{
			name: "Simple String Name",
			yamlData: `
id: "pkg:test"
type: pkg
name: simple-package
state: present
`,
			envDistro:    "ubuntu",
			expectedName: "simple-package",
			expectedID:   "pkg:test",
		},
		{
			name: "Map Name Match Distro",
			yamlData: `
id: "pkg:firefox"
type: pkg
name:
  arch: firefox
  ubuntu: firefox-browser
state: present
`,
			envDistro:    "ubuntu",
			expectedName: "firefox-browser",
			expectedID:   "pkg:firefox",
		},
		{
			name: "Map Name Fallback Default",
			yamlData: `
id: "pkg:chrome"
type: pkg
name:
  arch: google-chrome
  default: chromium
state: present
`,
			envDistro:    "fedora",
			expectedName: "chromium",
			expectedID:   "pkg:chrome",
		},
		{
			name: "Map Name No Match",
			yamlData: `
id: "pkg:noport"
type: pkg
name:
  arch: arch-only
state: present
`,
			envDistro:    "debian",
			expectedName: "", // Should be empty
			expectedID:   "pkg:noport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Env
			os.Setenv("VETO_DISTRO", tt.envDistro)
			defer os.Unsetenv("VETO_DISTRO")

			var res ResourceConfig
			err := yaml.Unmarshal([]byte(tt.yamlData), &res)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if res.Name != tt.expectedName {
				t.Errorf("Expected Name '%s', got '%s'", tt.expectedName, res.Name)
			}
			if res.ID != tt.expectedID {
				t.Errorf("Expected ID '%s', got '%s'", tt.expectedID, res.ID)
			}
		})
	}
}
