package docker

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ImageMetadata struct {
	Ref       string    `json:"ref"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
}

// getCacheDir returns the cache directory for aigogo
func getCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cache := filepath.Join(home, ".aigogo", "cache")
	if err := os.MkdirAll(cache, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cache, nil
}

// SanitizeImageRef converts an image reference to a safe filename
func SanitizeImageRef(ref string) string {
	// Replace special characters with underscores
	safe := strings.ReplaceAll(ref, "/", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	return safe
}

// Keep lowercase version for internal use
func sanitizeImageRef(ref string) string {
	return SanitizeImageRef(ref)
}

// parseImageRef parses a Docker image reference
// Format: [registry/]repository[:tag]
// Examples:
//   - docker.io/user/repo:tag
//   - ghcr.io/user/repo:v1.0.0
//   - user/repo:latest
func parseImageRef(ref string) (registry, repository, tag string, err error) {
	// Default values
	registry = "docker.io"
	tag = "latest"

	// Split by ':'
	parts := strings.Split(ref, ":")
	if len(parts) > 2 {
		return "", "", "", fmt.Errorf("invalid image reference: %s", ref)
	}

	if len(parts) == 2 {
		tag = parts[1]
	}

	// Parse registry and repository
	repoParts := strings.Split(parts[0], "/")

	if len(repoParts) >= 3 {
		// Has registry: registry/namespace/repo
		registry = repoParts[0]
		repository = strings.Join(repoParts[1:], "/")
	} else if len(repoParts) == 2 {
		// Could be registry/repo or namespace/repo
		// Check if first part looks like a registry (has . or :)
		if strings.Contains(repoParts[0], ".") || strings.Contains(repoParts[0], ":") {
			registry = repoParts[0]
			repository = repoParts[1]
		} else {
			// namespace/repo, use default registry
			repository = parts[0]
		}
	} else {
		// Just repo name
		repository = parts[0]
	}

	if repository == "" {
		return "", "", "", fmt.Errorf("invalid image reference: %s", ref)
	}

	return registry, repository, tag, nil
}

// calculateDigest calculates SHA256 digest of data (internal use)
func calculateDigest(data []byte) string {
	return CalculateDigest(data)
}

// CalculateDigest calculates SHA256 digest of data
func CalculateDigest(data []byte) string {
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:])
}

// CalculateDirectoryDigest calculates SHA256 digest of all files in a directory
// Files are processed in sorted order for deterministic results
func CalculateDirectoryDigest(dir string) (string, error) {
	h := sha256.New()

	// Collect all files
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort for deterministic hashing
	sortedFiles := make([]string, len(files))
	copy(sortedFiles, files)
	// Simple sort
	for i := 0; i < len(sortedFiles); i++ {
		for j := i + 1; j < len(sortedFiles); j++ {
			if sortedFiles[i] > sortedFiles[j] {
				sortedFiles[i], sortedFiles[j] = sortedFiles[j], sortedFiles[i]
			}
		}
	}

	// Hash each file
	for _, file := range sortedFiles {
		// Write filename
		h.Write([]byte(file))
		h.Write([]byte{0})

		// Write content
		content, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", file, err)
		}
		h.Write(content)
		h.Write([]byte{0})
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// getRegistryAPIEndpoint returns the actual API endpoint for a registry
// Docker Hub uses registry-1.docker.io for API, not docker.io
func getRegistryAPIEndpoint(registry string) string {
	if registry == "docker.io" {
		return "registry-1.docker.io"
	}
	return registry
}

// setAuthHeader sets the appropriate Authorization header for a registry.
// Docker Hub uses Bearer tokens (JWT from OAuth2 exchange).
// All other registries use Basic auth (base64 username:password).
func setAuthHeader(req *http.Request, registry, token string) {
	if token == "" {
		return
	}
	if registry == "docker.io" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("Authorization", "Basic "+token)
	}
}

// IsLocalReference checks if an image reference is local (no registry prefix)
func IsLocalReference(ref string) bool {
	// If no slash, it's local (e.g., "utils:1.0.0")
	if !strings.Contains(ref, "/") {
		return true
	}

	// Check if first part looks like a registry (has dot or known registry)
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) < 2 {
		return true
	}

	firstPart := parts[0]
	// Known registries or has a dot/colon (domain-like)
	if strings.Contains(firstPart, ".") ||
		strings.Contains(firstPart, ":") ||
		firstPart == "localhost" {
		return false // has registry
	}

	return true // looks like namespace/repo without registry
}

// ImageExistsInCache checks if an image exists in the local cache
// Checks both local build cache and registry pull cache
func ImageExistsInCache(ref string) bool {
	cacheDir, err := getCacheDir()
	if err != nil {
		return false
	}

	sanitized := sanitizeImageRef(ref)

	// Check local build cache first
	localPath := filepath.Join(cacheDir, sanitized)
	if _, err := os.Stat(localPath); err == nil {
		return true
	}

	// Check registry pull cache (images/ subdirectory)
	registryPath := filepath.Join(cacheDir, "images", sanitized)
	if _, err := os.Stat(registryPath); err == nil {
		return true
	}

	return false
}

// GetCachePath returns the local cache path for an image reference, or ""
// if the image is not in the cache. Checks local build cache first, then
// the registry pull cache.
func GetCachePath(ref string) string {
	cacheDir, err := getCacheDir()
	if err != nil {
		return ""
	}

	sanitized := sanitizeImageRef(ref)

	// Check local build cache first
	localPath := filepath.Join(cacheDir, sanitized)
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}

	// Check registry pull cache (images/ subdirectory)
	registryPath := filepath.Join(cacheDir, "images", sanitized)
	if _, err := os.Stat(registryPath); err == nil {
		return registryPath
	}

	return ""
}
