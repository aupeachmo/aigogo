package manifest

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindManifestDir searches for aigogo.json starting from the current directory
// and walking up the directory tree (like git does with .git).
// Returns the directory containing aigogo.json and the absolute path to the file.
func FindManifestDir() (string, string, error) {
	// Start from current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get home directory to prevent searching too high
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Search up the directory tree
	dir := currentDir
	for {
		manifestPath := filepath.Join(dir, "aigogo.json")

		// Check if aigogo.json exists in this directory
		if _, err := os.Stat(manifestPath); err == nil {
			return dir, manifestPath, nil
		}

		// Get parent directory
		parent := filepath.Dir(dir)

		// Stop if we've reached the root or home directory
		if parent == dir || parent == homeDir || parent == "/" {
			break
		}

		dir = parent
	}

	return "", "", fmt.Errorf("aigogo.json not found in current directory or any parent directory")
}

// FindManifest searches for and loads the manifest from the current or parent directories
func FindManifest() (*Manifest, string, error) {
	manifestDir, manifestPath, err := FindManifestDir()
	if err != nil {
		return nil, "", err
	}

	// Load the manifest
	manifest, err := Load(manifestPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load manifest from %s: %w", manifestPath, err)
	}

	return manifest, manifestDir, nil
}

// IncrementVersion increments the patch version of a semver string
// e.g., "0.1.1" -> "0.1.2", "1.2.3" -> "1.2.4"
func IncrementVersion(version string) (string, error) {
	var major, minor, patch int

	// Parse semver (simple version, doesn't handle pre-release or build metadata)
	n, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	if err != nil || n != 3 {
		return "", fmt.Errorf("invalid semver format: %s (expected X.Y.Z)", version)
	}

	// Increment patch version
	patch++

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}
