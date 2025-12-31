package resource

import (
	"testing"

	"github.com/melih-ucgun/veto/internal/core"
)

func TestCreateResourceWithParams(t *testing.T) {
	baseCtx := core.NewSystemContext(false)

	tests := []struct {
		name        string
		resType     string
		resName     string
		params      map[string]interface{}
		ctxOverride func(*core.SystemContext)
		wantErr     bool
	}{
		// ... Mevcut testler ...

		// YENİ: User & Group
		{
			name:    "Create User Resource",
			resType: "user",
			resName: "testuser",
			params: map[string]interface{}{
				"uid":    "1001",
				"shell":  "/bin/zsh",
				"groups": []string{"wheel", "docker"},
			},
			wantErr: false,
		},
		{
			name:    "Create Group Resource",
			resType: "group",
			resName: "docker",
			params: map[string]interface{}{
				"gid":    999,
				"system": true,
			},
			wantErr: false,
		},
		// YENİ: Template & LineInFile
		{
			name:    "Create Template Resource",
			resType: "template",
			resName: "/etc/config.conf",
			params: map[string]interface{}{
				"src":  "./templates/config.tmpl",
				"vars": map[string]interface{}{"Port": 8080},
			},
			wantErr: false,
		},
		{
			name:    "Create LineInFile Resource",
			resType: "line_in_file",
			resName: "/etc/hosts",
			params: map[string]interface{}{
				"line":   "127.0.0.1 localhost",
				"regexp": "^127\\.0\\.0\\.1",
			},
			wantErr: false,
		},
	}

	// Döngü mantığı aynı...
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localCtx := *baseCtx
			if tt.ctxOverride != nil {
				tt.ctxOverride(&localCtx)
			}
			res, err := CreateResourceWithParams(tt.resType, tt.resName, tt.params, &localCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && res == nil {
				t.Errorf("Returned nil resource")
			}
		})
	}
}
