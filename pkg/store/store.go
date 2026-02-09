package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Store manages the content-addressable storage for aigogo packages
type Store struct {
	rootDir string // ~/.aigogo/store
}

// StoredPackage represents a package stored in the CAS
type StoredPackage struct {
	Hash     string // Full SHA256 hash
	FilesDir string // Path to the files directory
	Manifest string // Path to the aigogo.json manifest
}

// NewStore creates a new Store instance with the default location (~/.aigogo/store)
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	rootDir := filepath.Join(home, ".aigogo", "store")
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	return &Store{rootDir: rootDir}, nil
}

// NewStoreAt creates a Store at a specific location (useful for testing)
func NewStoreAt(rootDir string) (*Store, error) {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}
	return &Store{rootDir: rootDir}, nil
}

// RootDir returns the store's root directory
func (s *Store) RootDir() string {
	return s.rootDir
}

// GetPath returns the full path for a given hash
// Structure: ~/.aigogo/store/sha256/ab/abcdef123.../
func (s *Store) GetPath(hash string) string {
	// Remove sha256: prefix if present
	hash = strings.TrimPrefix(hash, "sha256:")

	if len(hash) < 2 {
		return filepath.Join(s.rootDir, "sha256", hash)
	}

	// Use first 2 chars as subdirectory for better filesystem performance
	return filepath.Join(s.rootDir, "sha256", hash[:2], hash)
}

// Has checks if a hash exists in the store
func (s *Store) Has(hash string) bool {
	path := s.GetPath(hash)
	_, err := os.Stat(path)
	return err == nil
}

// Store stores files from a source directory into the CAS
// Returns the computed SHA256 hash of the contents
func (s *Store) Store(srcDir string, files []string, manifestData []byte) (string, error) {
	// Compute hash of all content
	hash, err := s.computeContentHash(srcDir, files, manifestData)
	if err != nil {
		return "", fmt.Errorf("failed to compute content hash: %w", err)
	}

	// Check if already stored
	if s.Has(hash) {
		return hash, nil
	}

	// Create storage directory
	storePath := s.GetPath(hash)
	filesDir := filepath.Join(storePath, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create store directory: %w", err)
	}

	// Copy files
	for _, file := range files {
		srcPath := filepath.Join(srcDir, file)
		dstPath := filepath.Join(filesDir, file)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			// Clean up on error
			_ = os.RemoveAll(storePath)
			return "", fmt.Errorf("failed to create directory for %s: %w", file, err)
		}

		// Copy file
		if err := copyFile(srcPath, dstPath); err != nil {
			_ = os.RemoveAll(storePath)
			return "", fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Write manifest
	manifestPath := filepath.Join(storePath, "aigogo.json")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		_ = os.RemoveAll(storePath)
		return "", fmt.Errorf("failed to write manifest: %w", err)
	}

	return hash, nil
}

// Get retrieves a stored package by hash
func (s *Store) Get(hash string) (*StoredPackage, error) {
	if !s.Has(hash) {
		return nil, fmt.Errorf("package not found in store: %s", hash)
	}

	storePath := s.GetPath(hash)
	return &StoredPackage{
		Hash:     hash,
		FilesDir: filepath.Join(storePath, "files"),
		Manifest: filepath.Join(storePath, "aigogo.json"),
	}, nil
}

// MakeReadOnly makes all files in a stored package read-only
func (s *Store) MakeReadOnly(hash string) error {
	pkg, err := s.Get(hash)
	if err != nil {
		return err
	}

	// Walk the files directory and set read-only permissions
	return filepath.Walk(pkg.FilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Directories need execute permission to be traversable
			return os.Chmod(path, 0555)
		}

		// Files get read-only
		return os.Chmod(path, 0444)
	})
}

// GetManifest reads and returns the manifest for a stored package
func (s *Store) GetManifest(hash string) (map[string]interface{}, error) {
	pkg, err := s.Get(hash)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(pkg.Manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return manifest, nil
}

// ListFiles returns all files in a stored package
func (s *Store) ListFiles(hash string) ([]string, error) {
	pkg, err := s.Get(hash)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.Walk(pkg.FilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(pkg.FilesDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	return files, err
}

// computeContentHash computes SHA256 hash of files and manifest
func (s *Store) computeContentHash(srcDir string, files []string, manifestData []byte) (string, error) {
	h := sha256.New()

	// Sort files for deterministic hashing
	sortedFiles := make([]string, len(files))
	copy(sortedFiles, files)
	sort.Strings(sortedFiles)

	// Hash each file's path and content
	for _, file := range sortedFiles {
		// Write filename to hash (for path integrity)
		h.Write([]byte(file))
		h.Write([]byte{0}) // null separator

		// Read and hash file content
		filePath := filepath.Join(srcDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", file, err)
		}
		h.Write(content)
		h.Write([]byte{0}) // null separator
	}

	// Also hash the manifest
	h.Write([]byte("__manifest__"))
	h.Write([]byte{0})
	h.Write(manifestData)

	return hex.EncodeToString(h.Sum(nil)), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		return err
	}

	if err := dstFile.Close(); err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// Delete removes a package from the store
func (s *Store) Delete(hash string) error {
	if !s.Has(hash) {
		return fmt.Errorf("package not found in store: %s", hash)
	}
	return os.RemoveAll(s.GetPath(hash))
}
