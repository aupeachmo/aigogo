package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aupeachmo/aigogo/pkg/manifest"
)

// LocalBuilder builds packages to local cache without pushing
type LocalBuilder struct {
	cacheDir string
}

// NewLocalBuilder creates a new local builder
func NewLocalBuilder() *LocalBuilder {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &LocalBuilder{
		cacheDir: filepath.Join(home, ".aigogo", "cache"),
	}
}

// BuildFromDir builds a package from a specific directory
func (b *LocalBuilder) BuildFromDir(srcDir string, imageRef string, m *manifest.Manifest, force bool) error {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to source directory
	if err := os.Chdir(srcDir); err != nil {
		return fmt.Errorf("failed to change to directory %s: %w", srcDir, err)
	}

	// Restore directory on exit
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Call the main Build function
	return b.Build(imageRef, m, force)
}

// Build builds a package to the local cache
func (b *LocalBuilder) Build(imageRef string, m *manifest.Manifest, force bool) error {
	// Normalize image reference
	normalizedRef := normalizeImageRef(imageRef)
	cacheKey := strings.ReplaceAll(normalizedRef, "/", "_")
	cacheKey = strings.ReplaceAll(cacheKey, ":", "_")

	imagePath := filepath.Join(b.cacheDir, cacheKey)

	// Check if already exists
	if !force {
		if _, err := os.Stat(imagePath); err == nil {
			return fmt.Errorf("package already exists in cache: %s\nUse --force to rebuild", imageRef)
		}
	}

	// Create cache directory
	if err := os.MkdirAll(b.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Clean existing if force rebuild
	if force {
		_ = os.RemoveAll(imagePath)
	}

	// Create image directory
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	fmt.Printf("Building to cache: %s\n", imagePath)

	// Generate dependency files if needed
	if m.Dependencies != nil && (len(m.Dependencies.Runtime) > 0 || len(m.Dependencies.Dev) > 0) {
		fmt.Println("Generating dependency files...")
		if err := generateDependencyFiles(m); err != nil {
			return fmt.Errorf("failed to generate dependency files: %w", err)
		}
	}

	// Discover files to include
	filesToCopy, err := discoverFiles(m)
	if err != nil {
		return fmt.Errorf("failed to discover files: %w", err)
	}

	if len(filesToCopy) == 0 {
		return fmt.Errorf("no files to package")
	}

	// Always include aigogo.json if it exists
	if _, err := os.Stat("aigogo.json"); err == nil {
		hasManifest := false
		for _, f := range filesToCopy {
			if f == "aigogo.json" {
				hasManifest = true
				break
			}
		}
		if !hasManifest {
			filesToCopy = append(filesToCopy, "aigogo.json")
		}
	}

	fmt.Printf("Packaging %d file(s)...\n", len(filesToCopy))

	// Copy files to cache
	for _, file := range filesToCopy {
		srcPath := file
		dstPath := filepath.Join(imagePath, file)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file, err)
		}

		// Read source file
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Write to destination
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", file, err)
		}

		fmt.Printf("  + %s\n", file)
	}

	// Save metadata
	metadata := LocalBuildMetadata{
		Name:     imageRef,
		Type:     "local-build",
		BuiltAt:  time.Now().Format(time.RFC3339),
		Source:   "local",
		Manifest: m,
	}

	metadataPath := filepath.Join(imagePath, ".aigogo-metadata.json")
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// LocalBuildMetadata stores information about local builds
type LocalBuildMetadata struct {
	Name     string             `json:"name"`
	Type     string             `json:"type"` // "local-build" or "registry-pull"
	BuiltAt  string             `json:"built_at"`
	Source   string             `json:"source"` // "local", registry URL, etc.
	Registry string             `json:"registry,omitempty"`
	Manifest *manifest.Manifest `json:"manifest,omitempty"`
}

// normalizeImageRef normalizes an image reference for consistent storage
func normalizeImageRef(ref string) string {
	// If no registry prefix, keep as-is
	if !strings.Contains(ref, "/") {
		return ref
	}

	// Check if first part looks like a registry (has a dot or is a known registry)
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) == 2 {
		firstPart := parts[0]
		// Known registries or has a dot (domain-like)
		if strings.Contains(firstPart, ".") ||
			firstPart == "docker" ||
			firstPart == "localhost" {
			return ref
		}
		// Otherwise treat as namespace/repo (no registry)
		return ref
	}

	return ref
}

// Helper functions

func generateDependencyFiles(m *manifest.Manifest) error {
	// Import depgen package logic here or call it
	// For now, placeholder - will be implemented when needed
	// TODO: Call depgen.GenerateDependencyFiles(m)
	return nil
}

func discoverFiles(m *manifest.Manifest) ([]string, error) {
	// Use the FileDiscovery from manifest package
	discovery, err := manifest.NewFileDiscovery(".", m.Files.Exclude)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize file discovery: %w", err)
	}
	return discovery.Discover(m.Files, m.Language)
}
