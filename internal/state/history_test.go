package state

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/melih-ucgun/veto/internal/types"
)

// MockRealFS implements state.FileSystem for tests using real os calls
type MockRealFS struct{}

func (fs *MockRealFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (fs *MockRealFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (fs *MockRealFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *MockRealFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (fs *MockRealFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func TestHistoryManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "veto-history-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "state.json")
	fs := &MockRealFS{}
	mgr, err := NewManager(stateFile, fs)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// 1. Add Transaction
	tx := types.Transaction{
		ID:        "tx1",
		Timestamp: time.Now(),
		Status:    "success",
		Changes: []types.TransactionChange{
			{Type: "pkg", Name: "vim", Action: "installed"},
		},
	}

	if err := mgr.AddTransaction(tx); err != nil {
		t.Fatalf("AddTransaction failed: %v", err)
	}

	// 2. Read Back
	// Create new manager instance to ensure reading from file
	mgr2, err := NewManager(stateFile, fs)
	if err != nil {
		t.Fatal(err)
	}

	txs := mgr2.GetTransactions()
	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	if txs[0].ID != "tx1" {
		t.Errorf("Transaction ID mismatch. Got %s, want tx1", txs[0].ID)
	}
	if len(txs[0].Changes) != 1 {
		t.Error("Transaction changes mismatch")
	}
}
