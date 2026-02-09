package docker

import (
	"fmt"
	"os"
	"path/filepath"
)

type Remover struct{}

func NewRemover() *Remover {
	return &Remover{}
}

// Remove deletes a cached image (both local builds and registry pulls)
func (r *Remover) Remove(imageRef string) error {
	cache, err := getCacheDir()
	if err != nil {
		return err
	}

	sanitized := sanitizeImageRef(imageRef)

	// Check for local build first (most common after our changes)
	localPath := filepath.Join(cache, sanitized)
	if _, err := os.Stat(localPath); err == nil {
		if err := os.RemoveAll(localPath); err != nil {
			return fmt.Errorf("failed to remove local build: %w", err)
		}
		return nil
	}

	// Check for registry pull
	registryPath := filepath.Join(cache, "images", sanitized)
	if _, err := os.Stat(registryPath); err == nil {
		if err := os.RemoveAll(registryPath); err != nil {
			return fmt.Errorf("failed to remove registry pull: %w", err)
		}
		return nil
	}

	return fmt.Errorf("image not found: %s", imageRef)
}

// RemoveAll deletes all cached packages (both local builds and registry pulls)
func (r *Remover) RemoveAll() error {
	cache, err := getCacheDir()
	if err != nil {
		return err
	}

	// Remove all contents of cache directory
	entries, err := os.ReadDir(cache)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(cache, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
		}
	}

	return nil
}
