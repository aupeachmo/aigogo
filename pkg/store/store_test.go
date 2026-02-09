package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStoreAt(t *testing.T) {
	tmpDir := t.TempDir()
	storeDir := filepath.Join(tmpDir, "test-store")

	s, err := NewStoreAt(storeDir)
	if err != nil {
		t.Fatalf("NewStoreAt failed: %v", err)
	}

	if s.RootDir() != storeDir {
		t.Errorf("RootDir = %q, want %q", s.RootDir(), storeDir)
	}

	// Verify directory was created
	if _, err := os.Stat(storeDir); os.IsNotExist(err) {
		t.Error("Store directory was not created")
	}
}

func TestStoreAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	storeDir := filepath.Join(tmpDir, "store")
	srcDir := filepath.Join(tmpDir, "src")

	// Create source files
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "test.py"), []byte("print('hello')"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "helper.py"), []byte("def help(): pass"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create store
	s, err := NewStoreAt(storeDir)
	if err != nil {
		t.Fatal(err)
	}

	// Store files
	manifest := []byte(`{"name": "test", "version": "1.0.0"}`)
	hash, err := s.Store(srcDir, []string{"test.py", "helper.py"}, manifest)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if hash == "" {
		t.Error("Store returned empty hash")
	}

	// Verify Has works
	if !s.Has(hash) {
		t.Error("Has returned false for stored package")
	}

	// Retrieve
	pkg, err := s.Get(hash)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if pkg.Hash != hash {
		t.Errorf("Package hash = %q, want %q", pkg.Hash, hash)
	}

	// Verify files exist
	testPath := filepath.Join(pkg.FilesDir, "test.py")
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Error("test.py not found in stored package")
	}

	// Verify manifest exists
	if _, err := os.Stat(pkg.Manifest); os.IsNotExist(err) {
		t.Error("Manifest not found in stored package")
	}
}

func TestStoreDeterministicHash(t *testing.T) {
	tmpDir := t.TempDir()
	storeDir := filepath.Join(tmpDir, "store")
	srcDir := filepath.Join(tmpDir, "src")

	// Create source files
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.py"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "b.py"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := NewStoreAt(storeDir)
	if err != nil {
		t.Fatal(err)
	}

	manifest := []byte(`{"name": "test"}`)

	// Store with files in one order
	hash1, err := s.Store(srcDir, []string{"a.py", "b.py"}, manifest)
	if err != nil {
		t.Fatal(err)
	}

	// Delete and store with files in different order
	_ = s.Delete(hash1)

	hash2, err := s.Store(srcDir, []string{"b.py", "a.py"}, manifest)
	if err != nil {
		t.Fatal(err)
	}

	// Hashes should be identical (files are sorted internally)
	if hash1 != hash2 {
		t.Errorf("Hashes differ: %q vs %q", hash1, hash2)
	}
}

func TestMakeReadOnly(t *testing.T) {
	tmpDir := t.TempDir()
	storeDir := filepath.Join(tmpDir, "store")
	srcDir := filepath.Join(tmpDir, "src")

	// Create source file
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "test.py"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := NewStoreAt(storeDir)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := s.Store(srcDir, []string{"test.py"}, []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	// Make read-only
	if err := s.MakeReadOnly(hash); err != nil {
		t.Fatalf("MakeReadOnly failed: %v", err)
	}

	// Try to write to the file - should fail
	pkg, _ := s.Get(hash)
	testPath := filepath.Join(pkg.FilesDir, "test.py")

	err = os.WriteFile(testPath, []byte("modified"), 0644)
	if err == nil {
		t.Error("Expected write to read-only file to fail")
	}

	// Restore permissions for cleanup
	_ = os.Chmod(pkg.FilesDir, 0755)
	_ = os.Chmod(testPath, 0644)
}

func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()
	storeDir := filepath.Join(tmpDir, "store")
	srcDir := filepath.Join(tmpDir, "src")

	// Create nested source files
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.py"), []byte("root"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "nested.py"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := NewStoreAt(storeDir)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := s.Store(srcDir, []string{"root.py", "subdir/nested.py"}, []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	files, err := s.ListFiles(hash)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("ListFiles returned %d files, want 2", len(files))
	}
}

func TestGetPath(t *testing.T) {
	s := &Store{rootDir: "/home/user/.aigogo/store"}

	tests := []struct {
		hash string
		want string
	}{
		{
			hash: "abc123def456",
			want: "/home/user/.aigogo/store/sha256/ab/abc123def456",
		},
		{
			hash: "sha256:abc123def456",
			want: "/home/user/.aigogo/store/sha256/ab/abc123def456",
		},
		{
			hash: "xy",
			want: "/home/user/.aigogo/store/sha256/xy/xy",
		},
	}

	for _, tt := range tests {
		got := s.GetPath(tt.hash)
		if got != tt.want {
			t.Errorf("GetPath(%q) = %q, want %q", tt.hash, got, tt.want)
		}
	}
}

func TestHasNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStoreAt(filepath.Join(tmpDir, "store"))
	if err != nil {
		t.Fatal(err)
	}

	if s.Has("nonexistent") {
		t.Error("Has returned true for non-existent hash")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	storeDir := filepath.Join(tmpDir, "store")
	srcDir := filepath.Join(tmpDir, "src")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "test.py"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := NewStoreAt(storeDir)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := s.Store(srcDir, []string{"test.py"}, []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	if !s.Has(hash) {
		t.Fatal("Package not stored")
	}

	if err := s.Delete(hash); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if s.Has(hash) {
		t.Error("Package still exists after delete")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStoreAt(filepath.Join(tmpDir, "store"))
	if err != nil {
		t.Fatal(err)
	}

	err = s.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent package")
	}
}
