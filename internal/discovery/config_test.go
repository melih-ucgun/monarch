package discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/melih-ucgun/veto/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransport mocks the transport layer
type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) Execute(ctx context.Context, command string) (string, error) {
	args := m.Called(ctx, command)
	return args.String(0), args.Error(1)
}

func (m *MockTransport) Upload(ctx context.Context, localPath, remotePath string) error {
	return nil
}

func (m *MockTransport) Download(ctx context.Context, remotePath, localPath string) error {
	return nil
}

func (m *MockTransport) Close() error {
	return nil
}

func TestDiscoverConfigs(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	sysCtx := &core.SystemContext{
		Context:   context.Background(),
		Transport: mockTransport,
		Cwd:       "/home/user",
	}

	// Create dummy files for fileExists check
	// We need to simulate file existence. 
	// Since DiscoverConfigs calls os.Stat, we actually need to create files on disk or mock os.Stat (hard in Go without interface).
	// Let's create a temporary directory structures.
	tmpDir := t.TempDir()
	sysCtx.Cwd = tmpDir // Fake home

    // Mock "home" dir env if needed or just use relative paths if function uses sysCtx.Cwd as home fallback
    
    // Create ~/.config/ghostty/config
    os.MkdirAll(filepath.Join(tmpDir, ".config", "ghostty"), 0755)
    os.WriteFile(filepath.Join(tmpDir, ".config", "ghostty", "config"), []byte(""), 0644)
    
    // Create /etc/nginx/nginx.conf (We can't easily write to /etc in test, 
    // but we can fake the ExpandPath logic or just focus on user configs for FS check if we can't write to /etc).
    // Actually, GetPackageFiles returns absolute paths. os.Stat will check absolute paths.
    // We cannot create /etc files in a normal test environment.
    // So testing the "Strategy 2" (System Configs) fully is hard without root/container.
    // BUT we can test "Strategy 3" (User Configs) easily.
    
    // For Strategy 2, code calls os.Stat on the returned file. 
    // If we return a file that exists in tmpDir from GetPackageFiles, it should work!
    
    fakeEtcFile := filepath.Join(tmpDir, "fake_etc_nginx.conf")
    os.WriteFile(fakeEtcFile, []byte(""), 0644)

	// Mock Transport for DetectManager -> pacman
	mockTransport.On("Execute", mock.Anything, "paru --version").Return("", assert.AnError)
	mockTransport.On("Execute", mock.Anything, "yay --version").Return("", assert.AnError)
	mockTransport.On("Execute", mock.Anything, "pacman --version").Return("Pacman v6.0.0", nil)

	// Mock Transport for GetPackageFiles -> pacman -Qlq nginx
	// We return our fake local file as if it was in /etc
    // Note: The logic in config.go expects prefix "/etc/". 
    // We can't fool it unless we actually create /etc/... OR we modify logic to be more flexible for test?
    // Or we just test Strategy 3 mainly.
    // Let's try to test Strategy 3 first.
    
    packages := []string{"ghostty"}
    
    // Act
    configs, err := DiscoverConfigs(sysCtx, packages)
    
    // Assert
    assert.NoError(t, err)
    assert.Len(t, configs, 1) // Should find ~/.config/ghostty/config
    if len(configs) > 0 {
        assert.Contains(t, configs[0].Path, "ghostty")
        assert.Equal(t, "ghostty", configs[0].PackageID)
    }
}
