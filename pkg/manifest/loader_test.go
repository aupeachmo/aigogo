package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "aigogo.json")

	content := `{
  "name": "test-package",
  "version": "1.0.0",
  "description": "A test package",
  "author": "Test Author",
  "language": {
    "name": "python",
    "version": ">=3.8"
  },
  "dependencies": {
    "runtime": [
      {"package": "requests", "version": ">=2.31.0"}
    ],
    "dev": [
      {"package": "pytest", "version": ">=7.0.0"}
    ]
  }
}`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if m.Name != "test-package" {
		t.Errorf("Name = %q, want %q", m.Name, "test-package")
	}
	if m.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", m.Version, "1.0.0")
	}
	if m.Language.Name != "python" {
		t.Errorf("Language.Name = %q, want %q", m.Language.Name, "python")
	}
	if len(m.Dependencies.Runtime) != 1 {
		t.Errorf("Runtime deps = %d, want 1", len(m.Dependencies.Runtime))
	}
	if len(m.Dependencies.Dev) != 1 {
		t.Errorf("Dev deps = %d, want 1", len(m.Dependencies.Dev))
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/aigogo.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "aigogo.json")

	if err := os.WriteFile(manifestPath, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(manifestPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "aigogo.json")

	// Missing required name field
	content := `{"version": "1.0.0", "language": {"name": "python"}}`
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(manifestPath)
	if err == nil {
		t.Error("Expected validation error for missing name")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "aigogo.json")

	m := &Manifest{
		Name:        "test-package",
		Version:     "1.0.0",
		Description: "A test package",
		Language:    Language{Name: "python", Version: ">=3.8"},
		Dependencies: &Dependencies{
			Runtime: []Dependency{
				{Package: "requests", Version: ">=2.31.0"},
			},
		},
	}

	if err := Save(manifestPath, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload and verify
	loaded, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load after Save failed: %v", err)
	}

	if loaded.Name != m.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, m.Name)
	}
	if loaded.Description != m.Description {
		t.Errorf("Description = %q, want %q", loaded.Description, m.Description)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		m       *Manifest
		wantErr bool
	}{
		{
			name: "valid manifest",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "python"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			m: &Manifest{
				Version:  "1.0.0",
				Language: Language{Name: "python"},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			m: &Manifest{
				Name:     "test",
				Language: Language{Name: "python"},
			},
			wantErr: true,
		},
		{
			name: "missing language name",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{},
			},
			wantErr: true,
		},
		{
			name: "unsupported language",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "cobol"},
			},
			wantErr: true,
		},
		{
			name: "deps without language version",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "python"},
				Dependencies: &Dependencies{
					Runtime: []Dependency{{Package: "requests", Version: ">=1.0"}},
				},
			},
			wantErr: true,
		},
		{
			name: "deps with language version",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "python", Version: ">=3.8"},
				Dependencies: &Dependencies{
					Runtime: []Dependency{{Package: "requests", Version: ">=1.0"}},
				},
			},
			wantErr: false,
		},
		{
			name: "dep missing package name",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "python", Version: ">=3.8"},
				Dependencies: &Dependencies{
					Runtime: []Dependency{{Version: ">=1.0"}},
				},
			},
			wantErr: true,
		},
		{
			name: "dep missing version",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "python", Version: ">=3.8"},
				Dependencies: &Dependencies{
					Runtime: []Dependency{{Package: "requests"}},
				},
			},
			wantErr: true,
		},
		{
			name: "dev dep missing package",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "python", Version: ">=3.8"},
				Dependencies: &Dependencies{
					Dev: []Dependency{{Version: ">=1.0"}},
				},
			},
			wantErr: true,
		},
		{
			name: "dev dep missing version",
			m: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Language: Language{Name: "python", Version: ">=3.8"},
				Dependencies: &Dependencies{
					Dev: []Dependency{{Package: "pytest"}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		lang  string
		valid bool
	}{
		{"python", true},
		{"javascript", true},
		{"go", true},
		{"rust", true},
		{"cobol", false},
		{"", false},
	}

	for _, tt := range tests {
		result := ValidateLanguage(tt.lang)
		if result != tt.valid {
			t.Errorf("ValidateLanguage(%q) = %v, want %v", tt.lang, result, tt.valid)
		}
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := SupportedLanguages()
	if len(langs) < 4 {
		t.Errorf("Expected at least 4 supported languages, got %d", len(langs))
	}

	expected := []string{"python", "javascript", "go", "rust"}
	for _, exp := range expected {
		found := false
		for _, lang := range langs {
			if lang == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %q in supported languages", exp)
		}
	}
}
