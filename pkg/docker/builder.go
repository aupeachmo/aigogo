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
)

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

// BuildImage creates a Docker image from scratch with the specified files
func (b *Builder) BuildImage(imageRef string, files []string, manifest interface{}) error {
	// Create a tar archive with the files
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add manifest as a special file
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Add manifest to tar
	if err := addToTar(tw, ".aigogo-manifest.json", manifestData, int64(len(manifestData))); err != nil {
		return err
	}

	// Add each file to the tar
	for _, file := range files {
		if err := addFileToTar(tw, file); err != nil {
			return fmt.Errorf("failed to add file %s: %w", file, err)
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Create the image structure
	// An OCI/Docker image is composed of:
	// 1. Config JSON (metadata about the image)
	// 2. Layer tar.gz files (filesystem changes)
	// 3. Manifest JSON (ties everything together)

	// For our use case, we'll create a minimal image structure
	// and save it to the local cache directory
	cache, err := getCacheDir()
	if err != nil {
		return err
	}

	imagePath := filepath.Join(cache, "images", sanitizeImageRef(imageRef))
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	// Save the layer (our tar archive)
	layerPath := filepath.Join(imagePath, "layer.tar")
	if err := os.WriteFile(layerPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write layer: %w", err)
	}

	// Create image metadata
	metadata := ImageMetadata{
		Ref:       imageRef,
		CreatedAt: time.Now(),
		Size:      int64(buf.Len()),
	}

	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metadataPath := filepath.Join(imagePath, "metadata.json")
	if err := os.WriteFile(metadataPath, metadataData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// BuildImageFromPath creates a Docker image from files in a specified directory
func (b *Builder) BuildImageFromPath(imageRef string, basePath string, files []string, manifest interface{}) error {
	// Create a tar archive with the files
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add manifest as a special file
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Add manifest to tar
	if err := addToTar(tw, ".aigogo-manifest.json", manifestData, int64(len(manifestData))); err != nil {
		return err
	}

	// Add each file to the tar from the base path
	for _, file := range files {
		fullPath := filepath.Join(basePath, file)
		if err := addFileToTarFromPath(tw, fullPath, file); err != nil {
			return fmt.Errorf("failed to add file %s: %w", file, err)
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Create the image structure and save to cache
	cache, err := getCacheDir()
	if err != nil {
		return err
	}

	imagePath := filepath.Join(cache, "images", sanitizeImageRef(imageRef))
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	// Save the layer
	layerPath := filepath.Join(imagePath, "layer.tar")
	if err := os.WriteFile(layerPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write layer: %w", err)
	}

	// Create image metadata
	metadata := ImageMetadata{
		Ref:       imageRef,
		CreatedAt: time.Now(),
		Size:      int64(buf.Len()),
	}

	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metadataPath := filepath.Join(imagePath, "metadata.json")
	if err := os.WriteFile(metadataPath, metadataData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func addFileToTar(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    filename,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

// addFileToTarFromPath adds a file to tar archive from a specific path with a given name in the archive
func addFileToTarFromPath(tw *tar.Writer, fullPath string, nameInArchive string) error {
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    nameInArchive,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

func addToTar(tw *tar.Writer, name string, data []byte, size int64) error {
	header := &tar.Header{
		Name:    name,
		Size:    size,
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err := tw.Write(data)
	return err
}
