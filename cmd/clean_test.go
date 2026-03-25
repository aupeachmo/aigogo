package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirStats(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty directory
	size, count := dirStats(tmpDir)
	if size != 0 {
		t.Errorf("dirStats empty dir: size = %d, want 0", size)
	}
	if count != 0 {
		t.Errorf("dirStats empty dir: count = %d, want 0", count)
	}

	// Non-existent directory
	size, count = dirStats("/nonexistent/path")
	if size != 0 || count != 0 {
		t.Errorf("dirStats non-existent: size=%d, count=%d, want 0,0", size, count)
	}

	// Directory with files
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	size, count = dirStats(tmpDir)
	if size != 10 {
		t.Errorf("dirStats with files: size = %d, want 10", size)
	}
	if count != 1 { // 1 top-level entry (sub/)
		t.Errorf("dirStats with files: count = %d, want 1", count)
	}
}

func TestCleanDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "test-clean")

	// Clean non-existent directory should not error
	if err := cleanDirectory(dir, "test"); err != nil {
		t.Errorf("cleanDirectory non-existent: unexpected error: %v", err)
	}

	// Create directory with content
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Clean should remove it
	if err := cleanDirectory(dir, "test"); err != nil {
		t.Errorf("cleanDirectory: unexpected error: %v", err)
	}

	// Verify removed
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("cleanDirectory did not remove the directory")
	}
}
