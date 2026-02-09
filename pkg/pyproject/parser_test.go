package pyproject

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPyProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Test not found
	_, err := FindPyProject(tmpDir)
	if err == nil {
		t.Error("Expected error when pyproject.toml not found")
	}

	// Create pyproject.toml
	pyprojectPath := filepath.Join(tmpDir, "pyproject.toml")
	if err := os.WriteFile(pyprojectPath, []byte("[project]\nname = \"test\""), 0644); err != nil {
		t.Fatal(err)
	}

	// Test found
	path, err := FindPyProject(tmpDir)
	if err != nil {
		t.Fatalf("FindPyProject failed: %v", err)
	}
	if path != pyprojectPath {
		t.Errorf("Path = %q, want %q", path, pyprojectPath)
	}
}

func TestParsePEP621(t *testing.T) {
	tmpDir := t.TempDir()
	pyprojectPath := filepath.Join(tmpDir, "pyproject.toml")

	content := `
[project]
name = "my-package"
version = "1.0.0"
description = "A test package"
requires-python = ">=3.8"
dependencies = [
    "requests>=2.31.0",
    "click~=8.1.0",
    "numpy>=1.20,<2.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.0.0",
    "black",
]
docs = [
    "sphinx>=4.0",
]
`
	if err := os.WriteFile(pyprojectPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pyproject, err := Parse(pyprojectPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if pyproject.Project == nil {
		t.Fatal("Project section is nil")
	}

	if pyproject.Project.Name != "my-package" {
		t.Errorf("Name = %q, want %q", pyproject.Project.Name, "my-package")
	}

	if pyproject.Project.RequiresPython != ">=3.8" {
		t.Errorf("RequiresPython = %q, want %q", pyproject.Project.RequiresPython, ">=3.8")
	}

	if len(pyproject.Project.Dependencies) != 3 {
		t.Errorf("Dependencies count = %d, want 3", len(pyproject.Project.Dependencies))
	}
}

func TestParsePoetry(t *testing.T) {
	tmpDir := t.TempDir()
	pyprojectPath := filepath.Join(tmpDir, "pyproject.toml")

	content := `
[tool.poetry]
name = "poetry-package"
version = "2.0.0"
description = "A Poetry package"

[tool.poetry.dependencies]
python = "^3.9"
requests = "^2.31.0"
flask = {version = "^2.0.0", optional = true}

[tool.poetry.dev-dependencies]
pytest = "^7.0.0"

[tool.poetry.group.test.dependencies]
coverage = "^7.0"
`
	if err := os.WriteFile(pyprojectPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pyproject, err := Parse(pyprojectPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if pyproject.Tool == nil || pyproject.Tool.Poetry == nil {
		t.Fatal("Poetry section is nil")
	}

	if pyproject.Tool.Poetry.Name != "poetry-package" {
		t.Errorf("Name = %q, want %q", pyproject.Tool.Poetry.Name, "poetry-package")
	}
}

func TestExtractDependenciesPEP621(t *testing.T) {
	pyproject := &PyProject{
		Project: &ProjectSection{
			Name:           "test-pkg",
			Version:        "1.0.0",
			RequiresPython: ">=3.8",
			Dependencies: []string{
				"requests>=2.31.0",
				"click~=8.1.0",
			},
			OptionalDeps: map[string][]string{
				"dev":  {"pytest>=7.0.0"},
				"test": {"coverage"},
			},
		},
	}

	result, err := ExtractDependencies(pyproject)
	if err != nil {
		t.Fatalf("ExtractDependencies failed: %v", err)
	}

	if result.Format != "pep621" {
		t.Errorf("Format = %q, want %q", result.Format, "pep621")
	}

	if result.PythonVersion != ">=3.8" {
		t.Errorf("PythonVersion = %q, want %q", result.PythonVersion, ">=3.8")
	}

	if len(result.Runtime) != 2 {
		t.Errorf("Runtime deps = %d, want 2", len(result.Runtime))
	}

	// Dev deps from optional-dependencies with "dev" or "test" in name
	if len(result.Dev) != 2 {
		t.Errorf("Dev deps = %d, want 2", len(result.Dev))
	}
}

func TestExtractDependenciesPoetry(t *testing.T) {
	pyproject := &PyProject{
		Tool: &ToolSection{
			Poetry: &PoetrySection{
				Name:    "poetry-pkg",
				Version: "1.0.0",
				Dependencies: map[string]interface{}{
					"python":   "^3.9",
					"requests": "^2.31.0",
					"flask":    map[string]interface{}{"version": "^2.0.0", "optional": true},
				},
				DevDeps: map[string]interface{}{
					"pytest": "^7.0.0",
				},
				Group: map[string]*DepGroup{
					"test": {
						Dependencies: map[string]interface{}{
							"coverage": "^7.0",
						},
					},
				},
			},
		},
	}

	result, err := ExtractDependencies(pyproject)
	if err != nil {
		t.Fatalf("ExtractDependencies failed: %v", err)
	}

	if result.Format != "poetry" {
		t.Errorf("Format = %q, want %q", result.Format, "poetry")
	}

	if result.PythonVersion != "^3.9" {
		t.Errorf("PythonVersion = %q, want %q", result.PythonVersion, "^3.9")
	}

	// Runtime: requests, flask (python is skipped)
	if len(result.Runtime) != 2 {
		t.Errorf("Runtime deps = %d, want 2", len(result.Runtime))
	}

	// Dev: pytest + coverage from test group
	if len(result.Dev) != 2 {
		t.Errorf("Dev deps = %d, want 2", len(result.Dev))
	}
}

func TestExtractDependenciesNoValidSection(t *testing.T) {
	pyproject := &PyProject{}

	_, err := ExtractDependencies(pyproject)
	if err == nil {
		t.Error("Expected error for pyproject without project or poetry section")
	}
}

func TestParsePEP508Dependency(t *testing.T) {
	tests := []struct {
		input   string
		pkg     string
		version string
	}{
		{"requests>=2.31.0", "requests", ">=2.31.0"},
		{"flask==2.0.0", "flask", "==2.0.0"},
		{"numpy>=1.20,<2.0", "numpy", ">=1.20,<2.0"},
		{"click~=8.1.0", "click", "~=8.1.0"},
		{"django>3.0", "django", ">3.0"},
		{"boto3<2.0", "boto3", "<2.0"},
		{"requests[security]>=2.31.0", "requests", ">=2.31.0"},
		{"simple-package", "simple-package", "*"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parsePEP508Dependency(tt.input)

			if tt.input == "" {
				if result != nil {
					t.Error("Expected nil for empty input")
				}
				return
			}

			if result == nil {
				t.Fatal("Result is nil")
			}

			if result.Package != tt.pkg {
				t.Errorf("Package = %q, want %q", result.Package, tt.pkg)
			}

			if result.Version != tt.version {
				t.Errorf("Version = %q, want %q", result.Version, tt.version)
			}
		})
	}
}

func TestParsePoetryDependency(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		value    interface{}
		expected string
		optional bool
	}{
		{"simple string", "requests", "^2.31.0", "^2.31.0", false},
		{"complex object", "flask", map[string]interface{}{"version": "^2.0.0"}, "^2.0.0", false},
		{"optional", "redis", map[string]interface{}{"version": "^4.0", "optional": true}, "^4.0", true},
		{"invalid", "bad", 123, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePoetryDependency(tt.pkg, tt.value)

			if tt.expected == "" {
				if result != nil {
					t.Error("Expected nil for invalid input")
				}
				return
			}

			if result == nil {
				t.Fatal("Result is nil")
			}

			if result.Package != tt.pkg {
				t.Errorf("Package = %q, want %q", result.Package, tt.pkg)
			}

			if result.Version != tt.expected {
				t.Errorf("Version = %q, want %q", result.Version, tt.expected)
			}

			if result.Optional != tt.optional {
				t.Errorf("Optional = %v, want %v", result.Optional, tt.optional)
			}
		})
	}
}

func TestNormalizeVersionConstraint(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"^3.9", "^3.9"},
		{map[string]interface{}{"version": ">=3.8"}, ">=3.8"},
		{map[string]interface{}{"other": "value"}, ""},
		{123, ""},
		{nil, ""},
	}

	for _, tt := range tests {
		result := normalizeVersionConstraint(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeVersionConstraint(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	pyprojectPath := filepath.Join(tmpDir, "pyproject.toml")

	// Invalid TOML
	if err := os.WriteFile(pyprojectPath, []byte("invalid = [toml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Parse(pyprojectPath)
	if err == nil {
		t.Error("Expected error for invalid TOML")
	}
}

func TestParseNonExistentFile(t *testing.T) {
	_, err := Parse("/nonexistent/path/pyproject.toml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
