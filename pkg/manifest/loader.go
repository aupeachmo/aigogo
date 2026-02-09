package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// Load reads and parses aigogo.json
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate required fields
	if err := Validate(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// Save writes manifest to file
func Save(path string, manifest *Manifest) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // Don't escape <, >, & for better readability
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(manifest); err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// Validate checks manifest for required fields and valid values
func Validate(m *Manifest) error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if m.Language.Name == "" {
		return fmt.Errorf("language.name is required")
	}
	if !ValidateLanguage(m.Language.Name) {
		return fmt.Errorf("unsupported language: %s (supported: %v)",
			m.Language.Name, SupportedLanguages())
	}
	if m.Dependencies != nil && m.Language.Version == "" {
		return fmt.Errorf("language.version is required when dependencies are specified")
	}

	// Validate dependencies
	if m.Dependencies != nil {
		for _, dep := range m.Dependencies.Runtime {
			if dep.Package == "" {
				return fmt.Errorf("dependency package name is required")
			}
			if dep.Version == "" {
				return fmt.Errorf("dependency version is required for %s", dep.Package)
			}
		}
		for _, dep := range m.Dependencies.Dev {
			if dep.Package == "" {
				return fmt.Errorf("dev dependency package name is required")
			}
			if dep.Version == "" {
				return fmt.Errorf("dev dependency version is required for %s", dep.Package)
			}
		}
	}

	return nil
}
