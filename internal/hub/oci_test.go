package hub

import (
	"testing"
)

func TestParseRef(t *testing.T) {
	tests := []struct {
		ref          string
		wantRegistry string
		wantRepo     string
		wantTag      string
		wantErr      bool
	}{
		{
			ref:          "oci://ghcr.io/melih/veto:v1",
			wantRegistry: "ghcr.io",
			wantRepo:     "melih/veto",
			wantTag:      "v1",
			wantErr:      false,
		},
		{
			ref:          "oci://docker.io/library/busybox:latest",
			wantRegistry: "registry-1.docker.io",
			wantRepo:     "library/busybox",
			wantTag:      "latest",
			wantErr:      false,
		},
		{
			ref:          "oci://my-registry.com/repo",
			wantRegistry: "my-registry.com",
			wantRepo:     "repo",
			wantTag:      "latest",
			wantErr:      false,
		},
		{
			ref:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			reg, repo, tag, err := parseRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if reg != tt.wantRegistry {
					t.Errorf("parseRef() registry = %v, want %v", reg, tt.wantRegistry)
				}
				if repo != tt.wantRepo {
					t.Errorf("parseRef() repo = %v, want %v", repo, tt.wantRepo)
				}
				if tag != tt.wantTag {
					t.Errorf("parseRef() tag = %v, want %v", tag, tt.wantTag)
				}
			}
		})
	}
}
