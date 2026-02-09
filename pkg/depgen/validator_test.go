package depgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aupeach/aigogo/pkg/manifest"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Error("NewValidator returned nil")
	}
}

func TestValidateAllDeclared(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	if err := os.WriteFile(pyFile, []byte("import requests\nimport flask"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language: manifest.Language{Name: "python"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "requests", Version: ">=2.31.0"},
				{Package: "flask", Version: ">=2.0.0"},
			},
		},
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{pyFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid result")
	}
	if len(result.MissingDeps) != 0 {
		t.Errorf("Expected no missing deps, got %v", result.MissingDeps)
	}
	if len(result.UnusedDeps) != 0 {
		t.Errorf("Expected no unused deps, got %v", result.UnusedDeps)
	}
}

func TestValidateMissingDeps(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	if err := os.WriteFile(pyFile, []byte("import requests\nimport flask\nimport django"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language: manifest.Language{Name: "python"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "requests", Version: ">=2.31.0"},
			},
		},
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{pyFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result due to missing deps")
	}

	if len(result.MissingDeps) != 2 {
		t.Errorf("Expected 2 missing deps, got %d: %v", len(result.MissingDeps), result.MissingDeps)
	}

	found := make(map[string]bool)
	for _, pkg := range result.MissingDeps {
		found[pkg] = true
	}

	if !found["flask"] || !found["django"] {
		t.Error("Expected flask and django to be missing")
	}
}

func TestValidateUnusedDeps(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	if err := os.WriteFile(pyFile, []byte("import requests"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language: manifest.Language{Name: "python"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "requests", Version: ">=2.31.0"},
				{Package: "flask", Version: ">=2.0.0"},
				{Package: "django", Version: ">=4.0.0"},
			},
		},
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{pyFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Unused deps are warnings, not errors
	if !result.Valid {
		t.Error("Unused deps should not make result invalid")
	}

	if len(result.UnusedDeps) != 2 {
		t.Errorf("Expected 2 unused deps, got %d: %v", len(result.UnusedDeps), result.UnusedDeps)
	}
}

func TestValidateNoDependenciesSection(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	if err := os.WriteFile(pyFile, []byte("import requests"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language:     manifest.Language{Name: "python"},
		Dependencies: nil,
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{pyFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid when imports exist but no deps declared")
	}

	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions for adding dependencies")
	}
}

func TestValidateNoDependenciesSectionNoImports(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	// Only stdlib imports
	if err := os.WriteFile(pyFile, []byte("import os\nimport sys"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language:     manifest.Language{Name: "python"},
		Dependencies: nil,
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{pyFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid when no external imports and no deps")
	}
}

func TestValidateVersionWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	if err := os.WriteFile(pyFile, []byte("import a\nimport b\nimport c"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language: manifest.Language{Name: "python"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "a", Version: "*"},       // No version
				{Package: "b", Version: "==1.0.0"}, // Exact version (Python)
				{Package: "c", Version: ">=1.0.0"}, // Good version
			},
		},
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{pyFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should have warnings for 'a' (no version) and 'b' (exact version)
	if len(result.Warnings) < 2 {
		t.Errorf("Expected at least 2 warnings, got %d", len(result.Warnings))
	}
}

func TestValidateJavaScriptExactVersion(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "test.js")

	if err := os.WriteFile(jsFile, []byte("import axios from 'axios'"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language: manifest.Language{Name: "javascript"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "axios", Version: "1.6.0"}, // Exact (no ^, ~, etc.)
			},
		},
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{jsFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	hasExactWarning := false
	for _, w := range result.Warnings {
		if contains(w, "exact version") {
			hasExactWarning = true
			break
		}
	}

	if !hasExactWarning {
		t.Error("Expected warning for exact version in JavaScript")
	}
}

func TestValidateHasNoVersion(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		version string
		noVer   bool
	}{
		{"", true},
		{"*", true},
		{"latest", true},
		{">=1.0.0", false},
		{"^1.0.0", false},
		{"1.0.0", false},
	}

	for _, tt := range tests {
		result := v.hasNoVersion(tt.version)
		if result != tt.noVer {
			t.Errorf("hasNoVersion(%q) = %v, want %v", tt.version, result, tt.noVer)
		}
	}
}

func TestValidateHasExactVersion(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		version string
		lang    string
		exact   bool
	}{
		// Python
		{"==1.0.0", "python", true},
		{">=1.0.0", "python", false},
		{"~=1.0.0", "python", false},
		// JavaScript
		{"1.0.0", "javascript", true},
		{"^1.0.0", "javascript", false},
		{"~1.0.0", "javascript", false},
		{">=1.0.0", "javascript", false},
		// Go (uses MVS, no exact)
		{"v1.0.0", "go", false},
		// Rust
		{"=1.0.0", "rust", true},
		{"1.0.0", "rust", false},
	}

	for _, tt := range tests {
		result := v.hasExactVersion(tt.version, tt.lang)
		if result != tt.exact {
			t.Errorf("hasExactVersion(%q, %q) = %v, want %v", tt.version, tt.lang, result, tt.exact)
		}
	}
}

func TestValidateSuggestDependency(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		pkg      string
		lang     string
		contains string
	}{
		{"requests", "python", ">=1.0.0,<2.0.0"},
		{"axios", "javascript", "^1.0.0"},
		{"errors", "go", "v1.0.0"},
		{"serde", "rust", "1.0"},
		{"pkg", "unknown", "1.0.0"},
	}

	for _, tt := range tests {
		result := v.suggestDependency(tt.pkg, tt.lang)
		if !contains(result, tt.pkg) {
			t.Errorf("suggestDependency(%q, %q) should contain package name", tt.pkg, tt.lang)
		}
		if !contains(result, tt.contains) {
			t.Errorf("suggestDependency(%q, %q) should contain %q", tt.pkg, tt.lang, tt.contains)
		}
	}
}

func TestValidateScanError(t *testing.T) {
	m := &manifest.Manifest{
		Language:     manifest.Language{Name: "python"},
		Dependencies: nil,
	}

	v := NewValidator()
	_, err := v.Validate(m, []string{"/nonexistent/file.py"})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestValidateImportsReturned(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	if err := os.WriteFile(pyFile, []byte("import requests\nimport flask"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Language:     manifest.Language{Name: "python"},
		Dependencies: nil,
	}

	v := NewValidator()
	result, err := v.Validate(m, []string{pyFile})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if len(result.Imports) != 2 {
		t.Errorf("Expected 2 imports in result, got %d", len(result.Imports))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
