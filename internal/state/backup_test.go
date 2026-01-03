package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "veto-backup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bm := NewBackupManager(tmpDir)
	txID := "tx1"

	// Create a dummy file to back up
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("original content")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Test CreateBackup
	backupPath, err := bm.CreateBackup(txID, srcFile)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file not created")
	}

	// 2. Test RestoreBackup
	// Modify source
	if err := os.WriteFile(srcFile, []byte("modified content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := bm.RestoreBackup(backupPath, srcFile); err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	restoredContent, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(restoredContent) != string(content) {
		t.Errorf("Restore failed. Got '%s', want '%s'", string(restoredContent), string(content))
	}

	// 3. Test CreateBackup non-existent file
	noFile := filepath.Join(tmpDir, "doesnotexist")
	path, err := bm.CreateBackup(txID, noFile)
	if err != nil {
		t.Errorf("Expected nil error for non-existent file, got %v", err)
	}
	if path != "" {
		t.Errorf("Expected empty path for non-existent file, got %s", path)
	}
}
