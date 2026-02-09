package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract extracts files from a cached image to a directory
func (e *Extractor) Extract(imageRef, outputDir string, force bool) ([]string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the image from local cache
	cache, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	// Check for local build first (direct in cache)
	localPath := filepath.Join(cache, sanitizeImageRef(imageRef))
	localMetadataPath := filepath.Join(localPath, ".aigogo-metadata.json")

	if _, err := os.Stat(localMetadataPath); err == nil {
		// It's a local build - extract from directory
		return e.extractFromDirectory(localPath, outputDir, force)
	}

	// Check for registry pull (in images/ subdirectory)
	imagePath := filepath.Join(cache, "images", sanitizeImageRef(imageRef))
	layerPath := filepath.Join(imagePath, "layer.tar")

	layerData, err := os.ReadFile(layerPath)
	if err != nil {
		return nil, fmt.Errorf("image not found locally: %w", err)
	}

	// Extract tar
	tr := tar.NewReader(bytes.NewReader(layerData))
	var extractedFiles []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Skip the manifest file
		if header.Name == ".aigogo-manifest.json" {
			continue
		}

		targetPath := filepath.Join(outputDir, header.Name)

		// Check if file exists
		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				return nil, fmt.Errorf("file already exists: %s (use -f to overwrite)", targetPath)
			}
		}

		// Create directory if needed
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		outFile, err := os.Create(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(outFile, tr); err != nil {
			_ = outFile.Close()
			return nil, fmt.Errorf("failed to extract file: %w", err)
		}
		if err := outFile.Close(); err != nil {
			return nil, fmt.Errorf("failed to close file: %w", err)
		}

		// Set permissions
		if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
			return nil, fmt.Errorf("failed to set permissions: %w", err)
		}

		extractedFiles = append(extractedFiles, targetPath)
	}

	return extractedFiles, nil
}

// extractFromDirectory extracts files from a local build directory
func (e *Extractor) extractFromDirectory(srcDir, dstDir string, force bool) ([]string, error) {
	var extractedFiles []string

	// Walk through the source directory
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip metadata file
		if filepath.Base(path) == ".aigogo-metadata.json" {
			return nil
		}

		// Get relative path from srcDir
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Check if file exists
		if !force {
			if _, err := os.Stat(dstPath); err == nil {
				return fmt.Errorf("file already exists: %s (use -f to overwrite)", dstPath)
			}
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}
		defer func() { _ = srcFile.Close() }()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		defer func() { _ = dstFile.Close() }()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}

		// Set permissions
		if err := os.Chmod(dstPath, info.Mode()); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}

		extractedFiles = append(extractedFiles, dstPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return extractedFiles, nil
}
