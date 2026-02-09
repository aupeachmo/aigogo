package lockfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	lock := New()

	if lock.Version != CurrentVersion {
		t.Errorf("Version = %d, want %d", lock.Version, CurrentVersion)
	}

	if lock.Packages == nil {
		t.Error("Packages map is nil")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "aigogo.lock")

	// Create lock file
	lock := New()
	lock.Add("my_utils", LockedPackage{
		Version:   "1.0.0",
		Integrity: "sha256:abc123",
		Source:    "docker.io/org/my-utils:1.0.0",
		Language:  "python",
		Files:     []string{"utils.py", "helper.py"},
	})

	// Save
	if err := Save(lockPath, lock); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := Load(lockPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != CurrentVersion {
		t.Errorf("Loaded version = %d, want %d", loaded.Version, CurrentVersion)
	}

	pkg, exists := loaded.Get("my_utils")
	if !exists {
		t.Fatal("Package my_utils not found")
	}

	if pkg.Version != "1.0.0" {
		t.Errorf("Package version = %q, want %q", pkg.Version, "1.0.0")
	}

	if pkg.Integrity != "sha256:abc123" {
		t.Errorf("Package integrity = %q, want %q", pkg.Integrity, "sha256:abc123")
	}

	if pkg.Language != "python" {
		t.Errorf("Package language = %q, want %q", pkg.Language, "python")
	}

	if len(pkg.Files) != 2 {
		t.Errorf("Package files count = %d, want 2", len(pkg.Files))
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/aigogo.lock")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestFindLockFile(t *testing.T) {
	// Create a directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(projectDir, "src", "pkg")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create lock file in project root
	lockPath := filepath.Join(projectDir, "aigogo.lock")
	lock := New()
	lock.Add("test_pkg", LockedPackage{
		Version:   "1.0.0",
		Integrity: "sha256:test",
		Source:    "test:1.0.0",
		Language:  "python",
	})
	if err := Save(lockPath, lock); err != nil {
		t.Fatal(err)
	}

	// Find from subdirectory
	foundPath, foundLock, err := FindLockFileFrom(subDir)
	if err != nil {
		t.Fatalf("FindLockFileFrom failed: %v", err)
	}

	if foundPath != lockPath {
		t.Errorf("Found path = %q, want %q", foundPath, lockPath)
	}

	if !foundLock.Has("test_pkg") {
		t.Error("Found lock file doesn't have test_pkg")
	}
}

func TestFindLockFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, err := FindLockFileFrom(tmpDir)
	if err == nil {
		t.Error("Expected error when lock file not found")
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-utils", "my_utils"},
		{"my.utils", "my_utils"},
		{"my-fancy.utils", "my_fancy_utils"},
		{"MyUtils", "MyUtils"},
		{"my_utils", "my_utils"},
		{"123pkg", "_123pkg"},
		{"pkg@special", "pkg_special"},
		{"", ""},
	}

	for _, tt := range tests {
		got := NormalizeName(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetPackageName(t *testing.T) {
	tests := []struct {
		imageRef string
		want     string
	}{
		{"docker.io/org/my-utils:1.0.0", "my-utils"},
		{"ghcr.io/user/pkg:latest", "pkg"},
		{"my-utils:1.0.0", "my-utils"},
		{"my-utils", "my-utils"},
		{"registry.example.com/namespace/deep/path/pkg:v1", "pkg"},
	}

	for _, tt := range tests {
		got := GetPackageName(tt.imageRef)
		if got != tt.want {
			t.Errorf("GetPackageName(%q) = %q, want %q", tt.imageRef, got, tt.want)
		}
	}
}

func TestLockFileAddRemoveHas(t *testing.T) {
	lock := New()

	// Test Has on empty
	if lock.Has("test") {
		t.Error("Has returned true on empty lock file")
	}

	// Add
	lock.Add("test", LockedPackage{Version: "1.0.0"})
	if !lock.Has("test") {
		t.Error("Has returned false after Add")
	}

	// Get
	pkg, exists := lock.Get("test")
	if !exists {
		t.Error("Get returned false after Add")
	}
	if pkg.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", pkg.Version, "1.0.0")
	}

	// Remove
	if !lock.Remove("test") {
		t.Error("Remove returned false for existing package")
	}
	if lock.Has("test") {
		t.Error("Has returned true after Remove")
	}

	// Remove non-existent
	if lock.Remove("nonexistent") {
		t.Error("Remove returned true for non-existent package")
	}
}

func TestGetIntegrityHash(t *testing.T) {
	tests := []struct {
		integrity string
		want      string
	}{
		{"sha256:abc123def", "abc123def"},
		{"abc123def", "abc123def"},
		{"sha256:", ""},
	}

	for _, tt := range tests {
		pkg := LockedPackage{Integrity: tt.integrity}
		got := pkg.GetIntegrityHash()
		if got != tt.want {
			t.Errorf("GetIntegrityHash() with %q = %q, want %q", tt.integrity, got, tt.want)
		}
	}
}

func TestLoadEmptyPackages(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "aigogo.lock")

	// Write minimal lock file
	content := `{"version": 1}`
	if err := os.WriteFile(lockPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lock, err := Load(lockPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if lock.Packages == nil {
		t.Error("Packages map should be initialized, not nil")
	}
}
