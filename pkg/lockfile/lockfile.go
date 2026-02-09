package lockfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// LockFileName is the name of the lock file
	LockFileName = "aigogo.lock"
	// CurrentVersion is the current lock file format version
	CurrentVersion = 1
)

// LockFile represents the aigogo.lock file
type LockFile struct {
	Version  int                      `json:"version"`
	Packages map[string]LockedPackage `json:"packages"`
}

// LockedPackage represents a single package entry in the lock file
type LockedPackage struct {
	Version   string   `json:"version"`
	Integrity string   `json:"integrity"` // sha256:...
	Source    string   `json:"source"`    // registry/repo:tag
	Language  string   `json:"language"`  // python|javascript
	Files     []string `json:"files"`
}

// New creates a new empty LockFile
func New() *LockFile {
	return &LockFile{
		Version:  CurrentVersion,
		Packages: make(map[string]LockedPackage),
	}
}

// Load reads and parses a lock file from the given path
func Load(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("lock file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	// Initialize map if nil (empty packages)
	if lock.Packages == nil {
		lock.Packages = make(map[string]LockedPackage)
	}

	return &lock, nil
}

// Save writes the lock file to the given path
func Save(path string, lock *LockFile) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	// Add trailing newline
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// FindLockFile searches for aigogo.lock starting from the current directory
// and walking up the directory tree (similar to how git finds .git)
func FindLockFile() (string, *LockFile, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return FindLockFileFrom(cwd)
}

// FindLockFileFrom searches for aigogo.lock starting from the given directory
func FindLockFileFrom(startDir string) (string, *LockFile, error) {
	dir := startDir

	for {
		lockPath := filepath.Join(dir, LockFileName)
		if _, err := os.Stat(lockPath); err == nil {
			lock, err := Load(lockPath)
			if err != nil {
				return "", nil, err
			}
			return lockPath, lock, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", nil, fmt.Errorf("aigogo.lock not found")
}

// NormalizeName converts a package name to a valid Python module name
// - Converts hyphens to underscores: my-utils -> my_utils
// - Converts dots to underscores: my.utils -> my_utils
// - Ensures it starts with a letter or underscore
func NormalizeName(name string) string {
	// Replace hyphens and dots with underscores
	normalized := strings.ReplaceAll(name, "-", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")

	// Ensure it starts with a letter or underscore
	if len(normalized) > 0 && !isValidIdentifierStart(normalized[0]) {
		normalized = "_" + normalized
	}

	// Remove any remaining invalid characters
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	normalized = re.ReplaceAllString(normalized, "_")

	return normalized
}

// isValidIdentifierStart checks if a byte is a valid start for a Python identifier
func isValidIdentifierStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

// GetPackageName extracts the package name from an image reference
// docker.io/org/my-utils:1.0.0 -> my-utils
func GetPackageName(imageRef string) string {
	// Remove registry prefix
	parts := strings.Split(imageRef, "/")
	repoTag := parts[len(parts)-1]

	// Remove tag
	if idx := strings.LastIndex(repoTag, ":"); idx != -1 {
		repoTag = repoTag[:idx]
	}

	return repoTag
}

// Add adds or updates a package in the lock file
func (l *LockFile) Add(name string, pkg LockedPackage) {
	l.Packages[name] = pkg
}

// Remove removes a package from the lock file
func (l *LockFile) Remove(name string) bool {
	if _, exists := l.Packages[name]; exists {
		delete(l.Packages, name)
		return true
	}
	return false
}

// Has checks if a package exists in the lock file
func (l *LockFile) Has(name string) bool {
	_, exists := l.Packages[name]
	return exists
}

// Get returns a package from the lock file
func (l *LockFile) Get(name string) (LockedPackage, bool) {
	pkg, exists := l.Packages[name]
	return pkg, exists
}

// GetIntegrityHash returns just the hash portion of the integrity string
// "sha256:abc123..." -> "abc123..."
func (p *LockedPackage) GetIntegrityHash() string {
	if strings.HasPrefix(p.Integrity, "sha256:") {
		return strings.TrimPrefix(p.Integrity, "sha256:")
	}
	return p.Integrity
}
