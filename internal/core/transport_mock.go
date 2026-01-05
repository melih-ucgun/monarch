package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// MockTransport implements Transport for testing purposes
type MockTransport struct {
	mu           sync.Mutex
	Expectations map[string]MockResponse
	Calls        []string
}

type MockResponse struct {
	Output string
	Error  error
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		Expectations: make(map[string]MockResponse),
		Calls:        make([]string, 0),
	}
}

func (m *MockTransport) Close() error {
	return nil
}

func (m *MockTransport) Execute(ctx context.Context, cmd string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, cmd)

	// Check for exact match
	if resp, ok := m.Expectations[cmd]; ok {
		return resp.Output, resp.Error
	}

	// Check for prefix/pattern match (simplified)
	for k, v := range m.Expectations {
		if strings.Contains(cmd, k) { // Basic matching
			return v.Output, v.Error
		}
	}

	return "", fmt.Errorf("unexpected command: %s", cmd)
}

func (m *MockTransport) CopyFile(ctx context.Context, localPath, remotePath string) error {
	return nil // No-op for now
}

func (m *MockTransport) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	return nil // No-op for now
}

func (m *MockTransport) GetFileSystem() FileSystem {
	return &RealFS{} // Or a mock FS if needed
}

func (m *MockTransport) GetOS(ctx context.Context) (string, error) {
	return "linux", nil
}

// Helpers for Test Setup

func (m *MockTransport) OnExecute(cmd string, output string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Expectations[cmd] = MockResponse{Output: output, Error: err}
}

func (m *MockTransport) AssertCalled(cmdFragment string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if strings.Contains(call, cmdFragment) {
			return true
		}
	}
	return false
}
