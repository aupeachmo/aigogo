package pyproject

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aupeachmo/aigogo/pkg/manifest"
)

// PyProject represents a minimal pyproject.toml structure
// supporting both Poetry and PEP 621 (uv) formats
type PyProject struct {
	// PEP 621 format (uv, setuptools, etc.)
	Project *ProjectSection `toml:"project"`

	// Poetry format
	Tool *ToolSection `toml:"tool"`
}

// ProjectSection represents the [project] section (PEP 621)
type ProjectSection struct {
	Name           string              `toml:"name"`
	Version        string              `toml:"version"`
	Description    string              `toml:"description"`
	RequiresPython string              `toml:"requires-python"`
	Dependencies   []string            `toml:"dependencies"`
	OptionalDeps   map[string][]string `toml:"optional-dependencies"`
}

// ToolSection represents the [tool] section
type ToolSection struct {
	Poetry *PoetrySection `toml:"poetry"`
}

// PoetrySection represents the [tool.poetry] section
type PoetrySection struct {
	Name         string                 `toml:"name"`
	Version      string                 `toml:"version"`
	Description  string                 `toml:"description"`
	Dependencies map[string]interface{} `toml:"dependencies"`
	DevDeps      map[string]interface{} `toml:"dev-dependencies"`
	Group        map[string]*DepGroup   `toml:"group"`
}

// DepGroup represents a poetry dependency group
type DepGroup struct {
	Dependencies map[string]interface{} `toml:"dependencies"`
}

// ParsedDependencies holds the extracted dependencies and metadata
type ParsedDependencies struct {
	Runtime       []manifest.Dependency
	Dev           []manifest.Dependency
	PythonVersion string
	Format        string // "poetry" or "pep621"
}

// FindPyProject searches for pyproject.toml in the given directory
func FindPyProject(startDir string) (string, error) {
	pyprojectPath := filepath.Join(startDir, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		return pyprojectPath, nil
	}
	return "", fmt.Errorf("pyproject.toml not found in %s", startDir)
}

// Parse reads and parses a pyproject.toml file
func Parse(path string) (*PyProject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read pyproject.toml: %w", err)
	}

	var pyproject PyProject
	if err := toml.Unmarshal(data, &pyproject); err != nil {
		return nil, fmt.Errorf("failed to parse pyproject.toml: %w", err)
	}

	return &pyproject, nil
}

// ExtractDependencies extracts dependencies from pyproject.toml
// Supports both Poetry and PEP 621 formats
func ExtractDependencies(pyproject *PyProject) (*ParsedDependencies, error) {
	result := &ParsedDependencies{
		Runtime: []manifest.Dependency{},
		Dev:     []manifest.Dependency{},
	}

	// Try PEP 621 format first (uv, modern)
	if pyproject.Project != nil {
		result.Format = "pep621"
		result.PythonVersion = pyproject.Project.RequiresPython

		// Parse runtime dependencies
		for _, dep := range pyproject.Project.Dependencies {
			pkgDep := parsePEP508Dependency(dep)
			if pkgDep != nil {
				result.Runtime = append(result.Runtime, *pkgDep)
			}
		}

		// Parse optional dependencies as dev dependencies
		// Note: PEP 621 doesn't distinguish dev deps, but optional-dependencies
		// often includes dev-related groups
		for group, deps := range pyproject.Project.OptionalDeps {
			isDev := strings.Contains(strings.ToLower(group), "dev") ||
				strings.Contains(strings.ToLower(group), "test") ||
				strings.Contains(strings.ToLower(group), "doc")

			for _, dep := range deps {
				pkgDep := parsePEP508Dependency(dep)
				if pkgDep != nil {
					if isDev {
						result.Dev = append(result.Dev, *pkgDep)
					} else {
						result.Runtime = append(result.Runtime, *pkgDep)
					}
				}
			}
		}

		return result, nil
	}

	// Try Poetry format
	if pyproject.Tool != nil && pyproject.Tool.Poetry != nil {
		result.Format = "poetry"

		// Extract Python version from dependencies
		if pythonVer, ok := pyproject.Tool.Poetry.Dependencies["python"]; ok {
			result.PythonVersion = normalizeVersionConstraint(pythonVer)
		}

		// Parse runtime dependencies
		for pkg, ver := range pyproject.Tool.Poetry.Dependencies {
			if pkg == "python" {
				continue // Skip python itself
			}
			dep := parsePoetryDependency(pkg, ver)
			if dep != nil {
				result.Runtime = append(result.Runtime, *dep)
			}
		}

		// Parse dev dependencies
		if pyproject.Tool.Poetry.DevDeps != nil {
			for pkg, ver := range pyproject.Tool.Poetry.DevDeps {
				dep := parsePoetryDependency(pkg, ver)
				if dep != nil {
					result.Dev = append(result.Dev, *dep)
				}
			}
		}

		// Parse dependency groups (modern Poetry)
		if pyproject.Tool.Poetry.Group != nil {
			for groupName, group := range pyproject.Tool.Poetry.Group {
				isDev := groupName == "dev" || groupName == "test" || groupName == "docs"
				for pkg, ver := range group.Dependencies {
					dep := parsePoetryDependency(pkg, ver)
					if dep != nil {
						if isDev {
							result.Dev = append(result.Dev, *dep)
						} else {
							result.Runtime = append(result.Runtime, *dep)
						}
					}
				}
			}
		}

		return result, nil
	}

	return nil, fmt.Errorf("pyproject.toml does not contain [project] (PEP 621) or [tool.poetry] sections")
}

// parsePEP508Dependency parses a PEP 508 dependency string
// Examples: "requests>=2.31.0", "flask==2.0.0", "numpy>=1.20,<2.0"
func parsePEP508Dependency(depStr string) *manifest.Dependency {
	depStr = strings.TrimSpace(depStr)
	if depStr == "" {
		return nil
	}

	// Split on common version specifiers
	var pkg, version string
	for _, sep := range []string{">=", "<=", "==", "~=", "!=", ">", "<"} {
		if idx := strings.Index(depStr, sep); idx != -1 {
			pkg = strings.TrimSpace(depStr[:idx])
			version = strings.TrimSpace(depStr[idx:])
			break
		}
	}

	// If no version specifier found, treat entire string as package name
	if pkg == "" {
		pkg = depStr
		version = "*"
	}

	// Handle extras like "requests[security]>=2.31.0"
	if idx := strings.Index(pkg, "["); idx != -1 {
		pkg = pkg[:idx]
	}

	return &manifest.Dependency{
		Package: pkg,
		Version: version,
	}
}

// parsePoetryDependency parses a Poetry dependency
// Can be a string (version) or a complex object
func parsePoetryDependency(pkg string, value interface{}) *manifest.Dependency {
	switch v := value.(type) {
	case string:
		// Simple version string: "^1.2.3" or ">=1.0,<2.0"
		return &manifest.Dependency{
			Package: pkg,
			Version: v,
		}
	case map[string]interface{}:
		// Complex dependency object
		// Example: { version = "^1.2.3", optional = true }
		if verStr, ok := v["version"].(string); ok {
			dep := &manifest.Dependency{
				Package: pkg,
				Version: verStr,
			}
			if optional, ok := v["optional"].(bool); ok {
				dep.Optional = optional
			}
			return dep
		}
	}
	return nil
}

// normalizeVersionConstraint converts interface{} to string version
func normalizeVersionConstraint(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]interface{}:
		if verStr, ok := v["version"].(string); ok {
			return verStr
		}
	}
	return ""
}
