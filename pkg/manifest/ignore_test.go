package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePattern(t *testing.T) {
	tests := []struct {
		input    string
		pattern  string
		negated  bool
		dirOnly  bool
		anchored bool
	}{
		{"*.pyc", "*.pyc", false, false, false},
		{"!important.log", "important.log", true, false, false},
		{"build/", "build", false, true, false},
		{"/root.txt", "root.txt", false, false, true},
		{"src/test/", "src/test", false, true, true},
		{"!src/keep/", "src/keep", true, true, true},
		{"**/*.log", "**/*.log", false, false, false},
		{"docs/**", "docs/**", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parsePattern(tt.input, 1)
			if p.pattern != tt.pattern {
				t.Errorf("pattern = %q, want %q", p.pattern, tt.pattern)
			}
			if p.negated != tt.negated {
				t.Errorf("negated = %v, want %v", p.negated, tt.negated)
			}
			if p.dirOnly != tt.dirOnly {
				t.Errorf("dirOnly = %v, want %v", p.dirOnly, tt.dirOnly)
			}
			if p.anchored != tt.anchored {
				t.Errorf("anchored = %v, want %v", p.anchored, tt.anchored)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		isDir   bool
		pattern string
		want    bool
	}{
		// Basic patterns
		{"simple match", "foo.pyc", false, "*.pyc", true},
		{"simple no match", "foo.py", false, "*.pyc", false},
		{"nested match", "src/foo.pyc", false, "*.pyc", true},
		{"deep nested match", "a/b/c/foo.pyc", false, "*.pyc", true},

		// Directory patterns
		{"dir pattern matches dir", "build", true, "build/", true},
		{"dir pattern no match file", "build", false, "build/", false},
		{"dir pattern nested", "src/build", true, "build/", true},

		// Anchored patterns
		{"anchored match", "src/test.py", false, "src/test.py", true},
		{"anchored no match nested", "foo/src/test.py", false, "src/test.py", false},
		{"anchored dir", "src/tests", true, "src/tests/", true},

		// Double star patterns
		{"doublestar all py", "test.py", false, "**/*.py", true},
		{"doublestar nested py", "src/test.py", false, "**/*.py", true},
		{"doublestar deep py", "a/b/c/test.py", false, "**/*.py", true},
		{"doublestar prefix", "docs/api/index.md", false, "docs/**", true},
		{"doublestar prefix file", "docs/readme.md", false, "docs/**", true},
		{"doublestar middle", "src/test/unit.py", false, "src/**/unit.py", true},
		{"doublestar middle deep", "src/a/b/c/unit.py", false, "src/**/unit.py", true},

		// Character class
		{"char class", "test1.py", false, "test[0-9].py", true},
		{"char class no match", "testa.py", false, "test[0-9].py", false},

		// Question mark
		{"question mark", "test1.py", false, "test?.py", true},
		{"question mark no match", "test12.py", false, "test?.py", false},

		// Exact directory name
		{"exact dir name", "node_modules", true, "node_modules", true},
		{"exact dir nested", "foo/node_modules", true, "node_modules", true},
		{"exact dir as file", "node_modules", false, "node_modules", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parsePattern(tt.pattern, 1)
			got := matchPattern(tt.path, tt.isDir, p)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %v, %q) = %v, want %v",
					tt.path, tt.isDir, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestIgnoreManagerNoFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "aigogo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	im, err := NewIgnoreManager(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if im.HasAigogoIgnore() {
		t.Error("HasAigogoIgnore() = true, want false")
	}

	// Default excludes should still work
	// Note: directory patterns like ".git/" and "node_modules/" only match directories
	// In practice, the file walker skips these directories entirely
	if !im.ShouldIgnore(".git", true) {
		t.Error(".git directory should be ignored by default")
	}
	if !im.ShouldIgnore("node_modules", true) {
		t.Error("node_modules directory should be ignored by default")
	}
	// File extension patterns should match files
	if !im.ShouldIgnore("cache.pyc", false) {
		t.Error("*.pyc files should be ignored by default")
	}
}

func TestIgnoreManagerWithFile(t *testing.T) {
	// Create temp directory with .aigogoignore
	tmpDir, err := os.MkdirTemp("", "aigogo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	ignoreContent := `# Test ignore file
*.log
*.tmp

# But keep important.log
!important.log

# Ignore build directory
build/

# Ignore all test files
**/*_test.go
`
	ignorePath := filepath.Join(tmpDir, ".aigogoignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	im, err := NewIgnoreManager(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !im.HasAigogoIgnore() {
		t.Error("HasAigogoIgnore() = false, want true")
	}

	tests := []struct {
		path  string
		isDir bool
		want  bool
		desc  string
	}{
		{"app.log", false, true, "*.log should be ignored"},
		{"logs/app.log", false, true, "nested *.log should be ignored"},
		{"important.log", false, false, "important.log should NOT be ignored (negated)"},
		{"cache.tmp", false, true, "*.tmp should be ignored"},
		{"build", true, true, "build/ directory should be ignored"},
		{"build", false, false, "build as file should NOT be ignored"},
		{"main_test.go", false, true, "test files should be ignored"},
		{"pkg/main_test.go", false, true, "nested test files should be ignored"},
		{"main.go", false, false, "regular go files should NOT be ignored"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := im.ShouldIgnore(tt.path, tt.isDir)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q, %v) = %v, want %v",
					tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestIgnoreManagerWithManifestExcludes(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "aigogo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .aigogoignore with negation
	ignoreContent := `# Un-ignore vendor that manifest excludes
!vendor/important.go
`
	ignorePath := filepath.Join(tmpDir, ".aigogoignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Manifest excludes vendor/
	manifestExcludes := []string{"vendor/**"}

	im, err := NewIgnoreManager(tmpDir, manifestExcludes)
	if err != nil {
		t.Fatal(err)
	}

	// vendor/foo.go should be ignored (from manifest)
	if !im.ShouldIgnore("vendor/foo.go", false) {
		t.Error("vendor/foo.go should be ignored by manifest exclude")
	}

	// vendor/important.go should NOT be ignored (negated in .aigogoignore)
	if im.ShouldIgnore("vendor/important.go", false) {
		t.Error("vendor/important.go should NOT be ignored (negated)")
	}
}

func TestShouldIgnoreWithReason(t *testing.T) {
	// Create temp directory with .aigogoignore
	tmpDir, err := os.MkdirTemp("", "aigogo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	ignoreContent := `*.log
`
	ignorePath := filepath.Join(tmpDir, ".aigogoignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	im, err := NewIgnoreManager(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	ignored, reason := im.ShouldIgnoreWithReason("app.log", false)
	if !ignored {
		t.Error("app.log should be ignored")
	}
	if reason == "" {
		t.Error("reason should not be empty")
	}
	// Reason should contain line number and pattern
	if !contains(reason, ".aigogoignore:1") {
		t.Errorf("reason should contain line number, got: %s", reason)
	}
}

func TestIgnoreManagerComments(t *testing.T) {
	// Create temp directory with .aigogoignore containing various comment styles
	tmpDir, err := os.MkdirTemp("", "aigogo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	ignoreContent := `# This is a comment
   # Indented comment
*.log

# Another comment
*.tmp
`
	ignorePath := filepath.Join(tmpDir, ".aigogoignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	im, err := NewIgnoreManager(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Comments should be ignored, patterns should work
	if !im.ShouldIgnore("app.log", false) {
		t.Error("*.log pattern should work despite comments")
	}
	if !im.ShouldIgnore("cache.tmp", false) {
		t.Error("*.tmp pattern should work despite comments")
	}
}

func TestIgnoreManagerEmptyLines(t *testing.T) {
	// Create temp directory with .aigogoignore containing empty lines
	tmpDir, err := os.MkdirTemp("", "aigogo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	ignoreContent := `*.log

*.tmp

`
	ignorePath := filepath.Join(tmpDir, ".aigogoignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	im, err := NewIgnoreManager(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Empty lines should be ignored, patterns should work
	if !im.ShouldIgnore("app.log", false) {
		t.Error("*.log pattern should work despite empty lines")
	}
	if !im.ShouldIgnore("cache.tmp", false) {
		t.Error("*.tmp pattern should work despite empty lines")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
