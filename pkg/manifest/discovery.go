package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileDiscovery handles automatic file discovery
type FileDiscovery struct {
	baseDir       string
	ignoreManager *IgnoreManager
}

// NewFileDiscovery creates a new file discovery instance
// manifestExcludes: patterns from files.exclude in aigogo.json
func NewFileDiscovery(baseDir string, manifestExcludes []string) (*FileDiscovery, error) {
	im, err := NewIgnoreManager(baseDir, manifestExcludes)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ignore manager: %w", err)
	}
	return &FileDiscovery{
		baseDir:       baseDir,
		ignoreManager: im,
	}, nil
}

// Discover finds files based on FileSpec
func (fd *FileDiscovery) Discover(spec FileSpec, language Language) ([]string, error) {
	patterns, isAuto := spec.GetIncludePatterns()

	if isAuto {
		// Auto-discovery based on language
		return fd.autoDiscover(language)
	}

	if len(patterns) == 0 {
		return nil, fmt.Errorf("no files specified and auto-discovery not enabled")
	}

	// Manual patterns
	return fd.discoverWithPatterns(patterns)
}

// autoDiscover finds files based on language conventions
func (fd *FileDiscovery) autoDiscover(language Language) ([]string, error) {
	patterns := fd.getLanguagePatterns(language.Name)
	return fd.discoverWithPatterns(patterns)
}

// GetIgnoreManager returns the ignore manager for external use (e.g., add command)
func (fd *FileDiscovery) GetIgnoreManager() *IgnoreManager {
	return fd.ignoreManager
}

// getLanguagePatterns returns file patterns for a language
func (fd *FileDiscovery) getLanguagePatterns(lang string) []string {
	patterns := map[string][]string{
		"python":     {"**/*.py"},
		"javascript": {"**/*.js", "**/*.ts", "**/*.jsx", "**/*.tsx", "**/*.mjs", "**/*.cjs"},
		"go":         {"**/*.go"},
		"rust":       {"**/*.rs"},
	}

	if p, ok := patterns[lang]; ok {
		return p
	}
	return []string{"**/*"}
}

// discoverWithPatterns finds files matching include patterns, using IgnoreManager for exclusions
func (fd *FileDiscovery) discoverWithPatterns(includePatterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	// Walk the directory
	err := filepath.Walk(fd.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(fd.baseDir, path)

		if info.IsDir() {
			// Check if directory should be excluded
			if relPath != "." && fd.ignoreManager.ShouldIgnore(relPath, true) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be excluded
		if fd.ignoreManager.ShouldIgnore(relPath, false) {
			return nil
		}

		// Check if file matches include patterns
		if fd.matchesAnyPattern(relPath, includePatterns) {
			if !seen[relPath] {
				files = append(files, relPath)
				seen[relPath] = true
			}
		}

		return nil
	})

	return files, err
}

// matchesAnyPattern checks if path matches any include pattern
func (fd *FileDiscovery) matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if fd.matchPattern(path, pattern) {
			return true
		}
	}
	return false
}

// matchPattern matches a path against a glob pattern
func (fd *FileDiscovery) matchPattern(path, pattern string) bool {
	// Use filepath.Match for simple patterns without **
	if !strings.Contains(pattern, "**") {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// For ** patterns, do custom matching
	// **/*.py should match both "file.py" and "dir/file.py" and "dir/subdir/file.py"
	if strings.HasPrefix(pattern, "**/") {
		// Match files in any directory (including root)
		suffix := strings.TrimPrefix(pattern, "**/")
		matched, _ := filepath.Match(suffix, path)
		if matched {
			return true
		}
		// Also try matching against the full path
		matched, _ = filepath.Match("*/"+suffix, path)
		if matched {
			return true
		}
		// For deeper nesting, check if path ends with matching part
		return strings.HasSuffix(path, "/"+suffix) || filepath.Base(path) == suffix ||
			(func() bool {
				m, _ := filepath.Match(suffix, filepath.Base(path))
				return m
			})()
	}

	return false
}
