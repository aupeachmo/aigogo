package manifest

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const AigogoIgnoreFile = ".aigogoignore"

// Pattern represents a parsed gitignore pattern
type Pattern struct {
	original   string // original line for error messages
	pattern    string // the pattern to match
	negated    bool   // starts with !
	dirOnly    bool   // ends with /
	anchored   bool   // contains / (not at end) - anchored to root
	lineNumber int    // line number in file for error messages
}

// IgnoreManager handles .aigogoignore and manifest exclude patterns
type IgnoreManager struct {
	baseDir         string
	patterns        []*Pattern
	hasAigogoIgnore bool
}

// NewIgnoreManager creates a new IgnoreManager
// baseDir: project root (where aigogo.json is located)
// manifestExcludes: patterns from files.exclude in aigogo.json
func NewIgnoreManager(baseDir string, manifestExcludes []string) (*IgnoreManager, error) {
	im := &IgnoreManager{
		baseDir:  baseDir,
		patterns: make([]*Pattern, 0),
	}

	// Add default excludes first (lowest priority)
	for _, p := range getDefaultExcludes() {
		im.patterns = append(im.patterns, parsePattern(p, 0))
	}

	// Add manifest excludes (medium priority)
	for _, p := range manifestExcludes {
		im.patterns = append(im.patterns, parsePattern(p, 0))
	}

	// Load .aigogoignore if it exists (highest priority)
	if err := im.loadAigogoIgnore(); err != nil {
		return nil, err
	}

	return im, nil
}

// loadAigogoIgnore loads patterns from .aigogoignore file if it exists
func (im *IgnoreManager) loadAigogoIgnore() error {
	ignorePath := filepath.Join(im.baseDir, AigogoIgnoreFile)

	file, err := os.Open(ignorePath)
	if os.IsNotExist(err) {
		return nil // No .aigogoignore file, that's fine
	}
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	im.hasAigogoIgnore = true

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		im.patterns = append(im.patterns, parsePattern(line, lineNum))
	}

	return scanner.Err()
}

// HasAigogoIgnore returns true if .aigogoignore file was found
func (im *IgnoreManager) HasAigogoIgnore() bool {
	return im.hasAigogoIgnore
}

// ShouldIgnore returns true if the path should be ignored
// path should be relative to baseDir, using forward slashes
// isDir should be true if the path is a directory
func (im *IgnoreManager) ShouldIgnore(path string, isDir bool) bool {
	ignored, _ := im.ShouldIgnoreWithReason(path, isDir)
	return ignored
}

// ShouldIgnoreWithReason returns whether path is ignored and the reason
func (im *IgnoreManager) ShouldIgnoreWithReason(path string, isDir bool) (bool, string) {
	// Normalize path to use forward slashes
	path = filepath.ToSlash(path)

	// Remove leading ./ if present
	path = strings.TrimPrefix(path, "./")

	ignored := false
	reason := ""

	// Process patterns in order - later patterns can override earlier ones
	for _, p := range im.patterns {
		if matchPattern(path, isDir, p) {
			if p.negated {
				ignored = false
				reason = ""
			} else {
				ignored = true
				if p.lineNumber > 0 {
					reason = ".aigogoignore:" + itoa(p.lineNumber) + ": " + p.original
				} else {
					reason = p.original
				}
			}
		}
	}

	return ignored, reason
}

// parsePattern parses a gitignore pattern line
func parsePattern(line string, lineNumber int) *Pattern {
	p := &Pattern{
		original:   line,
		lineNumber: lineNumber,
	}

	pattern := line

	// Handle trailing spaces (can be escaped with \)
	// For simplicity, just trim trailing spaces
	pattern = strings.TrimRight(pattern, " \t")

	// Check for negation
	if strings.HasPrefix(pattern, "!") {
		p.negated = true
		pattern = strings.TrimPrefix(pattern, "!")
	}

	// Check for directory only (trailing /)
	if strings.HasSuffix(pattern, "/") {
		p.dirOnly = true
		pattern = strings.TrimSuffix(pattern, "/")
	}

	// Check if anchored (contains / not at the end)
	// A pattern is anchored if it contains a slash (that's not trailing)
	// Exception: patterns starting with **/ are NOT anchored (they match at any depth)
	if strings.Contains(pattern, "/") && !strings.HasPrefix(pattern, "**/") {
		p.anchored = true
		// Remove leading / if present
		pattern = strings.TrimPrefix(pattern, "/")
	}

	p.pattern = pattern
	return p
}

// matchPattern checks if a path matches a pattern
func matchPattern(path string, isDir bool, p *Pattern) bool {
	// Directory-only patterns don't match files
	if p.dirOnly && !isDir {
		return false
	}

	pattern := p.pattern

	// Handle ** patterns
	if strings.Contains(pattern, "**") {
		return matchDoublestar(path, pattern)
	}

	if p.anchored {
		// Anchored patterns match from the root
		return matchGlob(path, pattern)
	}

	// Non-anchored patterns can match anywhere in the path
	// Try matching against the full path
	if matchGlob(path, pattern) {
		return true
	}

	// Try matching against each path component
	parts := strings.Split(path, "/")
	for i := range parts {
		subpath := strings.Join(parts[i:], "/")
		if matchGlob(subpath, pattern) {
			return true
		}
	}

	// Also try matching just the basename
	return matchGlob(filepath.Base(path), pattern)
}

// matchDoublestar handles patterns with **
func matchDoublestar(path, pattern string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 2 {
		prefix := parts[0]
		suffix := parts[1]

		// Remove leading/trailing slashes from suffix
		suffix = strings.TrimPrefix(suffix, "/")

		// **/ at start - matches in any directory
		if prefix == "" {
			// Match suffix against path or any subpath
			if suffix == "" {
				return true // ** matches everything
			}

			// Try matching suffix against path and all subpaths
			pathParts := strings.Split(path, "/")
			for i := range pathParts {
				subpath := strings.Join(pathParts[i:], "/")
				if matchGlob(subpath, suffix) {
					return true
				}
			}
			// Also match just the basename
			return matchGlob(filepath.Base(path), suffix)
		}

		// prefix/** at end - matches everything under prefix
		if suffix == "" || suffix == "/" {
			prefix = strings.TrimSuffix(prefix, "/")
			return strings.HasPrefix(path, prefix+"/") || path == prefix
		}

		// prefix/**/suffix - prefix must match start, suffix must match end
		prefix = strings.TrimSuffix(prefix, "/")
		if !strings.HasPrefix(path, prefix+"/") && path != prefix {
			return false
		}

		// Check if suffix matches the remaining path
		remaining := strings.TrimPrefix(path, prefix+"/")
		remainingParts := strings.Split(remaining, "/")
		for i := range remainingParts {
			subpath := strings.Join(remainingParts[i:], "/")
			if matchGlob(subpath, suffix) {
				return true
			}
		}
	}

	return false
}

// matchGlob matches a path against a glob pattern (supports *, ?, [...])
func matchGlob(path, pattern string) bool {
	// Use filepath.Match for basic glob matching
	// It handles *, ?, and [...] but not **
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	return matched
}

// getDefaultExcludes returns patterns to always exclude
func getDefaultExcludes() []string {
	return []string{
		// Version control (directories only)
		".git/",
		".svn/",
		".hg/",

		// Dependencies (directories only)
		"node_modules/",
		"venv/",
		".venv/",
		"env/",
		"__pycache__/",
		"*.pyc",
		".eggs/",
		"*.egg-info/",

		// Build artifacts (directories only, except binary extensions)
		"dist/",
		"build/",
		"target/",
		"*.so",
		"*.dylib",
		"*.dll",
		"*.exe",

		// IDE (directories only, except swap files)
		".vscode/",
		".idea/",
		"*.swp",
		"*.swo",
		".DS_Store",

		// aigogo
		".aigogo/",
		"aigogo.json",
	}
}

// itoa converts int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
