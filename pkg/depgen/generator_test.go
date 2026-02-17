package depgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Error("NewGenerator returned nil")
	}
}

func TestGenerateNoDependencies(t *testing.T) {
	g := NewGenerator()
	m := &manifest.Manifest{
		Name:         "test",
		Version:      "1.0.0",
		Language:     manifest.Language{Name: "python"},
		Dependencies: nil,
	}

	files, err := g.Generate(m, t.TempDir())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if files != nil {
		t.Errorf("Expected nil files, got %v", files)
	}
}

func TestGenerateUnsupportedLanguage(t *testing.T) {
	g := NewGenerator()
	m := &manifest.Manifest{
		Name:     "test",
		Version:  "1.0.0",
		Language: manifest.Language{Name: "cobol"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{{Package: "test", Version: "1.0"}},
		},
	}

	_, err := g.Generate(m, t.TempDir())
	if err == nil {
		t.Error("Expected error for unsupported language")
	}
}

func TestGeneratePython(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Name:        "my-package",
		Version:     "1.0.0",
		Description: "Test package",
		Author:      "Test Author",
		Language:    manifest.Language{Name: "python", Version: ">=3.8"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "requests", Version: ">=2.31.0"},
				{Package: "click", Version: "~=8.1.0"},
			},
			Dev: []manifest.Dependency{
				{Package: "pytest", Version: ">=7.0.0"},
			},
		},
	}

	files, err := g.Generate(m, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Check requirements.txt
	reqContent, err := os.ReadFile(filepath.Join(tmpDir, "requirements.txt"))
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}

	if !strings.Contains(string(reqContent), "requests>=2.31.0") {
		t.Error("requirements.txt missing requests")
	}
	if !strings.Contains(string(reqContent), "click~=8.1.0") {
		t.Error("requirements.txt missing click")
	}

	// Check pyproject.toml
	pyprojectContent, err := os.ReadFile(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("Failed to read pyproject.toml: %v", err)
	}

	pyStr := string(pyprojectContent)
	if !strings.Contains(pyStr, `name = "my-package"`) {
		t.Error("pyproject.toml missing name")
	}
	if !strings.Contains(pyStr, `requires-python = ">=3.8"`) {
		t.Error("pyproject.toml missing requires-python")
	}
	if !strings.Contains(pyStr, "[project.optional-dependencies]") {
		t.Error("pyproject.toml missing [project.optional-dependencies] section")
	}
	if !strings.Contains(pyStr, "aigogo = [") {
		t.Error("pyproject.toml missing aigogo group for runtime deps")
	}
	if !strings.Contains(pyStr, "requests>=2.31.0") {
		t.Error("pyproject.toml missing requests in aigogo group")
	}
	if !strings.Contains(pyStr, "aigogo-dev = [") {
		t.Error("pyproject.toml missing aigogo-dev group for dev deps")
	}
	if !strings.Contains(pyStr, "pytest>=7.0.0") {
		t.Error("pyproject.toml missing pytest in aigogo-dev group")
	}
	// Runtime deps should NOT be in a top-level dependencies section
	if strings.Contains(pyStr, "dependencies = [") {
		t.Error("pyproject.toml should not have top-level dependencies = [...], use [project.optional-dependencies] aigogo instead")
	}
}

func TestGenerateJavaScript(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Name:        "my-package",
		Version:     "1.0.0",
		Description: "Test package",
		Language:    manifest.Language{Name: "javascript", Version: ">=18.0.0"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "axios", Version: "^1.6.0"},
				{Package: "lodash", Version: "^4.17.21"},
			},
			Dev: []manifest.Dependency{
				{Package: "jest", Version: "^29.0.0"},
			},
		},
	}

	files, err := g.Generate(m, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "package.json"))
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}

	jsonStr := string(content)
	if !strings.Contains(jsonStr, `"name": "my-package"`) {
		t.Error("package.json missing name")
	}
	if !strings.Contains(jsonStr, `"axios": "^1.6.0"`) {
		t.Error("package.json missing axios")
	}
	if !strings.Contains(jsonStr, `"jest": "^29.0.0"`) {
		t.Error("package.json missing jest in devDependencies")
	}
	if !strings.Contains(jsonStr, `"node": ">=18.0.0"`) {
		t.Error("package.json missing engines.node")
	}
	// Check aigogo metadata
	if !strings.Contains(jsonStr, `"aigogo"`) {
		t.Error("package.json missing aigogo metadata key")
	}
	if !strings.Contains(jsonStr, `"managedDependencies"`) {
		t.Error("package.json missing managedDependencies in aigogo metadata")
	}
	if !strings.Contains(jsonStr, `"managedDevDependencies"`) {
		t.Error("package.json missing managedDevDependencies in aigogo metadata")
	}
}

func TestGenerateGo(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Name:     "github.com/user/pkg",
		Version:  "1.0.0",
		Language: manifest.Language{Name: "go", Version: "1.22"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "github.com/pkg/errors", Version: "v0.9.1"},
			},
		},
	}

	files, err := g.Generate(m, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	modStr := string(content)
	if !strings.Contains(modStr, "module github.com/user/pkg") {
		t.Error("go.mod missing module")
	}
	if !strings.Contains(modStr, "go 1.22") {
		t.Error("go.mod missing go version")
	}
	if !strings.Contains(modStr, "github.com/pkg/errors v0.9.1") {
		t.Error("go.mod missing dependency")
	}
}

func TestGenerateRust(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Name:        "my-crate",
		Version:     "1.0.0",
		Description: "A Rust crate",
		Language:    manifest.Language{Name: "rust", Version: "2021"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "serde", Version: "1.0"},
				{Package: "tokio", Version: "1.35"},
			},
			Dev: []manifest.Dependency{
				{Package: "criterion", Version: "0.5"},
			},
		},
	}

	files, err := g.Generate(m, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "Cargo.toml"))
	if err != nil {
		t.Fatalf("Failed to read Cargo.toml: %v", err)
	}

	cargoStr := string(content)
	if !strings.Contains(cargoStr, `name = "my-crate"`) {
		t.Error("Cargo.toml missing name")
	}
	if !strings.Contains(cargoStr, `edition = "2021"`) {
		t.Error("Cargo.toml missing edition")
	}
	if !strings.Contains(cargoStr, `serde = "1.0"`) {
		t.Error("Cargo.toml missing serde")
	}
	if !strings.Contains(cargoStr, "[dev-dependencies]") {
		t.Error("Cargo.toml missing dev-dependencies section")
	}
	if !strings.Contains(cargoStr, `criterion = "0.5"`) {
		t.Error("Cargo.toml missing criterion")
	}
}

func TestGeneratePythonNoDevDeps(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Name:     "simple-pkg",
		Version:  "1.0.0",
		Language: manifest.Language{Name: "python", Version: ">=3.8"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "requests", Version: ">=2.31.0"},
			},
		},
	}

	_, err := g.Generate(m, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatal(err)
	}

	pyStr := string(content)
	// Should have optional-dependencies with aigogo group (runtime deps)
	if !strings.Contains(pyStr, "[project.optional-dependencies]") {
		t.Error("pyproject.toml should have [project.optional-dependencies] for aigogo runtime deps")
	}
	if !strings.Contains(pyStr, "aigogo = [") {
		t.Error("pyproject.toml should have aigogo group for runtime deps")
	}
	// Should NOT have aigogo-dev group
	if strings.Contains(pyStr, "aigogo-dev = [") {
		t.Error("pyproject.toml should not have aigogo-dev group when no dev deps")
	}
}

func TestGenerateJavaScriptNoDevDeps(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	m := &manifest.Manifest{
		Name:     "simple-pkg",
		Version:  "1.0.0",
		Language: manifest.Language{Name: "javascript"},
		Dependencies: &manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Package: "axios", Version: "^1.0.0"},
			},
		},
	}

	_, err := g.Generate(m, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "package.json"))
	if err != nil {
		t.Fatal(err)
	}

	jsonStr := string(content)
	if strings.Contains(jsonStr, "devDependencies") {
		t.Error("package.json should not have devDependencies when none specified")
	}
	if strings.Contains(jsonStr, "managedDevDependencies") {
		t.Error("package.json should not have managedDevDependencies when no dev deps")
	}
	// Should still have aigogo metadata with managedDependencies
	if !strings.Contains(jsonStr, `"aigogo"`) {
		t.Error("package.json should have aigogo metadata key")
	}
	if !strings.Contains(jsonStr, `"managedDependencies"`) {
		t.Error("package.json should have managedDependencies in aigogo metadata")
	}
}
