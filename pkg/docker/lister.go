package docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aupeachmo/aigogo/pkg/manifest"
)

type Lister struct{}

func NewLister() *Lister {
	return &Lister{}
}

// CachedImage represents a cached image with its metadata
type CachedImage struct {
	Name      string
	Type      string // "local-build" or "registry-pull"
	Source    string
	BuildTime time.Time
	Size      int64
	Manifest  *manifest.Manifest // The aigogo.json manifest if available
}

// List returns all cached images with their metadata
func (l *Lister) ListDetailed() ([]CachedImage, error) {
	cache, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	var images []CachedImage

	// First, scan local builds in cache/ directory
	entries, err := os.ReadDir(cache)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip the "images" subdirectory - we'll handle it separately
		if entry.Name() == "images" {
			continue
		}

		imageDir := filepath.Join(cache, entry.Name())

		// Check for local build metadata
		localMetadataPath := filepath.Join(imageDir, ".aigogo-metadata.json")
		if data, err := os.ReadFile(localMetadataPath); err == nil {
			var metadata LocalBuildMetadata
			if err := json.Unmarshal(data, &metadata); err == nil {
				buildTime, _ := time.Parse(time.RFC3339, metadata.BuiltAt)

				// Calculate directory size
				size, _ := getDirSize(imageDir)

				// Try to load aigogo.json manifest
				var aigogoManifest *manifest.Manifest
				aigogoPath := filepath.Join(imageDir, "aigogo.json")
				if manifestData, err := os.ReadFile(aigogoPath); err == nil {
					var m manifest.Manifest
					if err := json.Unmarshal(manifestData, &m); err == nil {
						aigogoManifest = &m
					}
				}

				images = append(images, CachedImage{
					Name:      metadata.Name,
					Type:      "local-build",
					Source:    "local",
					BuildTime: buildTime,
					Size:      size,
					Manifest:  aigogoManifest,
				})
			}
		}
	}

	// Second, scan registry pulls in cache/images/ subdirectory
	imagesDir := filepath.Join(cache, "images")
	if entries, err := os.ReadDir(imagesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			imageDir := filepath.Join(imagesDir, entry.Name())

			// Check for registry pull metadata
			registryMetadataPath := filepath.Join(imageDir, "metadata.json")
			if data, err := os.ReadFile(registryMetadataPath); err == nil {
				var metadata ImageMetadata
				if err := json.Unmarshal(data, &metadata); err == nil {
					// Calculate directory size
					size, _ := getDirSize(imageDir)

					// Try to load aigogo.json manifest from layer.tar if it exists
					var aigogoManifest *manifest.Manifest
					layerPath := filepath.Join(imageDir, "layer.tar")
					if layerData, err := os.ReadFile(layerPath); err == nil {
						// Try to extract aigogo.json from the tar
						if manifestData := extractManifestFromTar(layerData); manifestData != nil {
							var m manifest.Manifest
							if err := json.Unmarshal(manifestData, &m); err == nil {
								aigogoManifest = &m
							}
						}
					}

					images = append(images, CachedImage{
						Name:      metadata.Ref,
						Type:      "registry-pull",
						Source:    "registry",
						BuildTime: metadata.CreatedAt,
						Size:      size,
						Manifest:  aigogoManifest,
					})
				}
			}
		}
	}

	return images, nil
}

// List returns simple list of cached image names (for backward compatibility)
func (l *Lister) List() ([]string, error) {
	detailed, err := l.ListDetailed()
	if err != nil {
		return nil, err
	}

	var images []string
	for _, img := range detailed {
		images = append(images, img.Name)
	}

	return images, nil
}

// getDirSize calculates total size of a directory
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// extractManifestFromTar extracts aigogo.json from a tar archive
func extractManifestFromTar(tarData []byte) []byte {
	tr := tar.NewReader(bytes.NewReader(tarData))

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}

		// Look for aigogo.json in the tar
		if header.Name == "aigogo.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil
			}
			return data
		}
	}

	return nil
}
