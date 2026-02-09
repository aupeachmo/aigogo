package depgen

import (
	"fmt"
	"strings"

	"github.com/aupeach/aigogo/pkg/manifest"
)

// ValidationResult contains validation results
type ValidationResult struct {
	Valid       bool
	Errors      []string
	Warnings    []string
	Suggestions []string
	MissingDeps []string
	UnusedDeps  []string
	Imports     []ImportInfo
}

// ImportInfo represents a detected import
type ImportInfo struct {
	Package    string
	SourceFile string
	LineNumber int
}

// Validator validates dependencies
type Validator struct {
	scanner *Scanner
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		scanner: NewScanner(),
	}
}

// Validate checks if declared dependencies match actual imports
func (v *Validator) Validate(m *manifest.Manifest, files []string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
		MissingDeps: []string{},
		UnusedDeps:  []string{},
	}

	// Scan files for imports
	imports, err := v.scanner.ScanFiles(files, m.Language.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	result.Imports = imports

	if m.Dependencies == nil {
		if len(imports) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("Found %d imports but no dependencies declared", len(imports)))

			result.Suggestions = append(result.Suggestions, "Add dependencies section to aigogo.json:")
			for _, imp := range imports {
				result.Suggestions = append(result.Suggestions,
					v.suggestDependency(imp.Package, m.Language.Name))
			}
		}
		return result, nil
	}

	// Build maps
	imported := make(map[string]bool)
	for _, imp := range imports {
		imported[imp.Package] = true
	}

	declared := make(map[string]bool)
	for _, dep := range m.Dependencies.Runtime {
		declared[dep.Package] = true
	}

	// Find missing dependencies
	for pkg := range imported {
		if !declared[pkg] {
			result.MissingDeps = append(result.MissingDeps, pkg)
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Package '%s' is imported but not declared", pkg))
		}
	}

	// Find unused dependencies
	for pkg := range declared {
		if !imported[pkg] {
			result.UnusedDeps = append(result.UnusedDeps, pkg)
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Package '%s' is declared but not imported", pkg))
		}
	}

	// Check version constraints
	for _, dep := range m.Dependencies.Runtime {
		if v.hasNoVersion(dep.Version) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Package '%s' has no version constraint", dep.Package))
		}
		if v.hasExactVersion(dep.Version, m.Language.Name) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Package '%s' uses exact version (may cause conflicts)", dep.Package))
		}
	}

	// Suggest missing dependencies
	if len(result.MissingDeps) > 0 {
		result.Suggestions = append(result.Suggestions, "Add missing dependencies:")
		for _, pkg := range result.MissingDeps {
			result.Suggestions = append(result.Suggestions,
				v.suggestDependency(pkg, m.Language.Name))
		}
	}

	// Set validity
	if len(result.MissingDeps) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("Missing %d required dependencies", len(result.MissingDeps)))
	}

	return result, nil
}

func (v *Validator) hasNoVersion(version string) bool {
	version = strings.TrimSpace(version)
	return version == "" || version == "*" || version == "latest"
}

func (v *Validator) hasExactVersion(version, lang string) bool {
	switch lang {
	case "python":
		return strings.HasPrefix(version, "==")
	case "javascript":
		return !strings.ContainsAny(version, "^~><")
	case "go":
		return false // Go uses minimum version selection
	case "rust":
		return strings.HasPrefix(version, "=")
	}
	return false
}

func (v *Validator) suggestDependency(pkg, lang string) string {
	switch lang {
	case "python":
		return fmt.Sprintf(`  {"package": "%s", "version": ">=1.0.0,<2.0.0"}`, pkg)
	case "javascript":
		return fmt.Sprintf(`  {"package": "%s", "version": "^1.0.0"}`, pkg)
	case "go":
		return fmt.Sprintf(`  {"package": "%s", "version": "v1.0.0"}`, pkg)
	case "rust":
		return fmt.Sprintf(`  {"package": "%s", "version": "1.0"}`, pkg)
	}
	return fmt.Sprintf(`  {"package": "%s", "version": "1.0.0"}`, pkg)
}
