package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindManifestDirInCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "aigogo.json")

	content := `{"name": "test", "version": "1.0.0", "language": {"name": "python"}}`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	dir, path, err := FindManifestDir()
	if err != nil {
		t.Fatalf("FindManifestDir failed: %v", err)
	}

	if dir != tmpDir {
		t.Errorf("Dir = %q, want %q", dir, tmpDir)
	}
	if path != manifestPath {
		t.Errorf("Path = %q, want %q", path, manifestPath)
	}
}

func TestFindManifestDirInParent(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(tmpDir, "aigogo.json")
	content := `{"name": "test", "version": "1.0.0", "language": {"name": "python"}}`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to nested subdir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	dir, path, err := FindManifestDir()
	if err != nil {
		t.Fatalf("FindManifestDir failed: %v", err)
	}

	if dir != tmpDir {
		t.Errorf("Dir = %q, want %q", dir, tmpDir)
	}
	if path != manifestPath {
		t.Errorf("Path = %q, want %q", path, manifestPath)
	}
}

func TestFindManifestDirNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to empty dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	_, _, err := FindManifestDir()
	if err == nil {
		t.Error("Expected error when aigogo.json not found")
	}
}

func TestFindManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "aigogo.json")

	content := `{"name": "test-pkg", "version": "2.0.0", "language": {"name": "go"}}`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	m, dir, err := FindManifest()
	if err != nil {
		t.Fatalf("FindManifest failed: %v", err)
	}

	if m.Name != "test-pkg" {
		t.Errorf("Name = %q, want %q", m.Name, "test-pkg")
	}
	if m.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", m.Version, "2.0.0")
	}
	if dir != tmpDir {
		t.Errorf("Dir = %q, want %q", dir, tmpDir)
	}
}

func TestFindManifestInvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "aigogo.json")

	// Invalid: missing required fields
	content := `{"description": "no name or version"}`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, _, err := FindManifest()
	if err == nil {
		t.Error("Expected error for invalid manifest")
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"0.0.1", "0.0.2", false},
		{"1.0.0", "1.0.1", false},
		{"1.2.3", "1.2.4", false},
		{"0.1.9", "0.1.10", false},
		{"10.20.30", "10.20.31", false},
		{"invalid", "", true},
		{"1.0", "", true},
		{"1", "", true},
		{"", "", true},
		{"a.b.c", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := IncrementVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("IncrementVersion(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}
