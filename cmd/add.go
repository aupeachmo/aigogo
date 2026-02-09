package cmd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/docker"
	"github.com/aupeachmo/aigogo/pkg/lockfile"
	"github.com/aupeachmo/aigogo/pkg/manifest"
	"github.com/aupeachmo/aigogo/pkg/pyproject"
	"github.com/aupeachmo/aigogo/pkg/store"
)

func addCmd() *Command {
	return &Command{
		Name:        "add",
		Description: "Add packages, files, or dependencies",
		Run: func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: aigogo add <package-ref|file|dep|dev> [args...]\n\nSubcommands:\n  <registry/repo:tag>         Add a package to aigogo.lock\n  file <path>...              Add files to include list\n  dep <pkg> <ver>             Add runtime dependency\n  dep --from-pyproject        Import all dependencies from pyproject.toml\n  dev <pkg> <ver>             Add development dependency\n  dev --from-pyproject        Import dev dependencies from pyproject.toml\n\nExamples:\n  aigogo add docker.io/org/my-utils:1.0.0\n  aigogo add file utils.py helpers.py\n  aigogo add dep requests >=2.28.0")
			}

			subcommand := args[0]
			subArgs := args[1:]

			switch subcommand {
			case "file":
				return addFiles(subArgs)
			case "dep":
				return addDependencyCmd(subArgs, false)
			case "dev":
				return addDependencyCmd(subArgs, true)
			default:
				// If not a known subcommand, treat as package reference
				if looksLikePackageRef(subcommand) {
					return addPackage(subcommand)
				}
				return fmt.Errorf("unknown subcommand '%s'\nValid subcommands: file, dep, dev\nOr provide a package reference like: docker.io/org/package:tag", subcommand)
			}
		},
	}
}

// looksLikePackageRef checks if the argument looks like a package reference
// Package refs contain "/" or ":" (e.g., docker.io/org/pkg:1.0.0 or pkg:1.0.0)
func looksLikePackageRef(arg string) bool {
	return strings.Contains(arg, "/") || strings.Contains(arg, ":")
}

// addPackage adds a package to the lock file via CAS
func addPackage(imageRef string) error {
	fmt.Printf("Adding package: %s\n\n", imageRef)

	// Check local cache first before pulling from registry
	var srcDir string
	var relFiles []string
	var needCleanup bool

	if cachePath := docker.GetCachePath(imageRef); cachePath != "" {
		fmt.Println("Found in local cache...")
		srcDir = cachePath

		// Collect files from cache (skip metadata)
		entries, err := os.ReadDir(cachePath)
		if err != nil {
			return fmt.Errorf("failed to read local cache: %w", err)
		}
		for _, entry := range entries {
			if entry.Name() == ".aigogo-metadata.json" {
				continue
			}
			if entry.IsDir() {
				subFiles, err := collectFiles(filepath.Join(cachePath, entry.Name()), entry.Name())
				if err != nil {
					return err
				}
				relFiles = append(relFiles, subFiles...)
			} else {
				relFiles = append(relFiles, entry.Name())
			}
		}
	} else {
		// Pull from registry
		fmt.Println("Pulling from registry...")
		puller := docker.NewPuller()
		if err := puller.Pull(imageRef); err != nil {
			return fmt.Errorf("failed to pull package: %w", err)
		}

		// Extract to temp directory
		tmpDir, err := os.MkdirTemp("", "aigogo-add-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		srcDir = tmpDir
		needCleanup = true

		fmt.Println("Extracting package...")
		extractor := docker.NewExtractor()
		extractedFiles, err := extractor.Extract(imageRef, tmpDir, true)
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return fmt.Errorf("failed to extract package: %w", err)
		}

		// Convert to relative paths
		for _, f := range extractedFiles {
			relPath, err := filepath.Rel(tmpDir, f)
			if err != nil {
				_ = os.RemoveAll(tmpDir)
				return err
			}
			relFiles = append(relFiles, relPath)
		}
	}

	if needCleanup {
		defer func() { _ = os.RemoveAll(srcDir) }()
	}

	// Read manifest to get metadata
	manifestPath := filepath.Join(srcDir, "aigogo.json")
	var pkgManifest *manifest.Manifest
	var manifestData []byte

	if data, err := os.ReadFile(manifestPath); err == nil {
		manifestData = data
		var m manifest.Manifest
		if err := json.Unmarshal(data, &m); err == nil {
			pkgManifest = &m
		}
	}

	// Determine package info
	pkgName := lockfile.GetPackageName(imageRef)
	pkgVersion := "unknown"
	pkgLanguage := "python" // default

	if pkgManifest != nil {
		if pkgManifest.Name != "" {
			pkgName = pkgManifest.Name
		}
		if pkgManifest.Version != "" {
			pkgVersion = pkgManifest.Version
		}
		if pkgManifest.Language.Name != "" {
			pkgLanguage = strings.ToLower(pkgManifest.Language.Name)
		}
	} else {
		// Extract version from tag if no manifest
		if idx := strings.LastIndex(imageRef, ":"); idx != -1 {
			pkgVersion = imageRef[idx+1:]
		}
		// Create minimal manifest
		manifestData, _ = json.MarshalIndent(map[string]interface{}{
			"name":     pkgName,
			"version":  pkgVersion,
			"language": map[string]string{"name": pkgLanguage},
		}, "", "  ")
	}

	// Store in CAS
	fmt.Println("Storing in content-addressable store...")
	cas, err := store.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	hash, err := cas.Store(srcDir, relFiles, manifestData)
	if err != nil {
		return fmt.Errorf("failed to store package: %w", err)
	}

	// Make read-only
	if err := cas.MakeReadOnly(hash); err != nil {
		fmt.Printf("⚠ Warning: failed to make files read-only: %v\n", err)
	}

	// Find or create lock file
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	lockPath, lock, err := lockfile.FindLockFileFrom(cwd)
	if err != nil {
		// Create new lock file in current directory
		lockPath = filepath.Join(cwd, lockfile.LockFileName)
		lock = lockfile.New()
	}

	// Add package to lock file
	// Normalize name for Python (hyphens → underscores); keep original for JS
	lockName := pkgName
	if pkgLanguage == "python" {
		lockName = lockfile.NormalizeName(pkgName)
	}
	lock.Add(lockName, lockfile.LockedPackage{
		Version:   pkgVersion,
		Integrity: "sha256:" + hash,
		Source:    imageRef,
		Language:  pkgLanguage,
		Files:     relFiles,
	})

	// Save lock file
	if err := lockfile.Save(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save lock file: %w", err)
	}

	fmt.Printf("\n✓ Added %s@%s to %s\n", pkgName, pkgVersion, lockPath)
	fmt.Printf("  Hash: sha256:%s\n", hash[:16]+"...")
	fmt.Printf("  Files: %d\n", len(relFiles))
	fmt.Printf("  Language: %s\n", pkgLanguage)

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'aigogo install' to create import links")
	fmt.Println("  2. Commit aigogo.lock to version control")

	// Show import hint
	switch pkgLanguage {
	case "python":
		fmt.Printf("\nImport with: from aigogo.%s import ...\n", lockName)
	case "javascript", "typescript":
		fmt.Printf("\nImport with: import ... from '@aigogo/%s'\n", pkgName)
	}

	return nil
}

// collectFiles recursively collects file paths from a directory, returning
// paths relative to the given prefix.
func collectFiles(dir string, prefix string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		relPath := prefix + "/" + entry.Name()
		if entry.IsDir() {
			subFiles, err := collectFiles(filepath.Join(dir, entry.Name()), relPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			files = append(files, relPath)
		}
	}
	return files, nil
}

func addFiles(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("add file", flag.ContinueOnError)
	force := fs.Bool("force", false, "Add files even if they match ignore patterns")

	// Separate flags from positional args so flags can appear anywhere
	// (e.g., "add file utils.py --force" works the same as "add file --force utils.py")
	var flagArgs []string
	var posArgs []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flagArgs = append(flagArgs, args[i])
			// Check if next arg is the flag value (doesn't start with -)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			posArgs = append(posArgs, args[i])
		}
	}

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	posArgs = append(posArgs, fs.Args()...)
	args = posArgs

	// Find and load manifest (supports subdirectories)
	m, manifestDir, err := manifest.FindManifest()
	if err != nil {
		return fmt.Errorf("failed to find aigogo.json: %w\nRun 'aigogo init' first", err)
	}

	manifestPath := filepath.Join(manifestDir, "aigogo.json")

	// Create ignore manager to check if files are ignored
	ignoreManager, err := manifest.NewIgnoreManager(manifestDir, m.Files.Exclude)
	if err != nil {
		return fmt.Errorf("failed to load ignore patterns: %w", err)
	}

	// Check if files.include is "auto"
	if str, ok := m.Files.Include.(string); ok && str == "auto" {
		return fmt.Errorf("files.include is set to 'auto'\nPlease edit aigogo.json to change it to an array before adding files")
	}

	// Get file paths
	var filePaths []string
	if len(args) > 0 {
		filePaths = args
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("File paths (space-separated): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read file paths: %w", err)
		}
		filePaths = strings.Fields(strings.TrimSpace(input))
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("at least one file path is required")
	}

	// Get existing patterns
	existingPatterns, _ := m.Files.GetIncludePatterns()
	if existingPatterns == nil {
		existingPatterns = []string{}
	}

	// Validate and collect new files
	var addedFiles []string
	for _, path := range filePaths {
		// Check if already exists
		found := false
		for _, existing := range existingPatterns {
			if existing == path {
				fmt.Printf("⚠ Skipping '%s' (already in include list)\n", path)
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Check if it's a glob pattern
		isGlob := strings.Contains(path, "*") || strings.Contains(path, "?") || strings.Contains(path, "[")

		// Check if file/pattern is ignored (unless --force is used)
		if !*force && !isGlob {
			// For literal paths, check against ignore manager
			info, statErr := os.Stat(path)
			isDir := statErr == nil && info.IsDir()
			if ignored, reason := ignoreManager.ShouldIgnoreWithReason(path, isDir); ignored {
				fmt.Printf("⚠ Skipping '%s' (ignored by %s)\n", path, reason)
				fmt.Println("  Use --force to add anyway")
				continue
			}
		}

		if isGlob {
			// For glob patterns, check if at least one file matches
			matches, err := filepath.Glob(path)
			if err != nil {
				return fmt.Errorf("invalid glob pattern '%s': %w", path, err)
			}
			if len(matches) == 0 {
				return fmt.Errorf("glob pattern '%s' matches no files", path)
			}
			// For globs, warn about any ignored files that would match
			if !*force {
				var ignoredCount int
				for _, match := range matches {
					info, _ := os.Stat(match)
					isDir := info != nil && info.IsDir()
					if ignoreManager.ShouldIgnore(match, isDir) {
						ignoredCount++
					}
				}
				if ignoredCount > 0 {
					fmt.Printf("⚠ Note: %d file(s) matching '%s' are ignored by .aigogoignore\n", ignoredCount, path)
				}
			}
			addedFiles = append(addedFiles, path)
		} else {
			// For literal paths, check if file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", path)
			}
			addedFiles = append(addedFiles, path)
		}
	}

	if len(addedFiles) == 0 {
		fmt.Println("No new files to add")
		return nil
	}

	// Update include list
	existingPatterns = append(existingPatterns, addedFiles...)
	m.Files.Include = existingPatterns

	// Save updated manifest
	if err := manifest.Save(manifestPath, m); err != nil {
		return fmt.Errorf("failed to save aigogo.json: %w", err)
	}

	fmt.Printf("✓ Added %d file(s) to include list:\n", len(addedFiles))
	for _, f := range addedFiles {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}

func addDependencyCmd(args []string, isDev bool) error {
	// Parse flags
	fs := flag.NewFlagSet("add dep", flag.ContinueOnError)
	fromPyproject := fs.Bool("from-pyproject", false, "Import dependencies from pyproject.toml")

	// Separate flags from positional args so flags can appear anywhere
	var flagArgs []string
	var posArgs []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flagArgs = append(flagArgs, args[i])
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			posArgs = append(posArgs, args[i])
		}
	}

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	posArgs = append(posArgs, fs.Args()...)

	// Check if --from-pyproject flag is set
	if *fromPyproject {
		return addDependenciesFromPyproject(isDev)
	}

	// Otherwise, use manual dependency addition
	return addDependency(posArgs, isDev)
}

func addDependency(args []string, isDev bool) error {
	depType := "runtime"
	if isDev {
		depType = "development"
	}

	// Find and load manifest (supports subdirectories)
	m, manifestDir, err := manifest.FindManifest()
	if err != nil {
		return fmt.Errorf("failed to find aigogo.json: %w\nRun 'aigogo init' first", err)
	}

	manifestPath := filepath.Join(manifestDir, "aigogo.json")

	reader := bufio.NewReader(os.Stdin)

	// Get package name
	var pkgName string
	if len(args) > 0 {
		pkgName = args[0]
	} else {
		fmt.Print("Package name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read package name: %w", err)
		}
		pkgName = strings.TrimSpace(input)
	}

	if pkgName == "" {
		return fmt.Errorf("package name is required")
	}

	// Initialize dependencies if nil
	if m.Dependencies == nil {
		m.Dependencies = &manifest.Dependencies{
			Runtime: []manifest.Dependency{},
			Dev:     []manifest.Dependency{},
		}
	}

	// Check if already exists
	targetList := m.Dependencies.Runtime
	if isDev {
		targetList = m.Dependencies.Dev
	}
	for _, dep := range targetList {
		if dep.Package == pkgName {
			return fmt.Errorf("package '%s' is already declared as a %s dependency with version '%s'", pkgName, depType, dep.Version)
		}
	}

	// Get version
	var version string
	if len(args) > 1 {
		version = args[1]
	} else {
		// Show suggested version format based on language
		suggestion := suggestVersionFormat(m.Language.Name, pkgName)
		fmt.Printf("Version constraint %s: ", suggestion)

		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read version: %w", err)
		}
		version = strings.TrimSpace(input)
	}

	if version == "" {
		return fmt.Errorf("version is required")
	}

	// Add the dependency
	newDep := manifest.Dependency{
		Package: pkgName,
		Version: version,
	}

	if isDev {
		m.Dependencies.Dev = append(m.Dependencies.Dev, newDep)
	} else {
		m.Dependencies.Runtime = append(m.Dependencies.Runtime, newDep)
	}

	// Save updated manifest
	if err := manifest.Save(manifestPath, m); err != nil {
		return fmt.Errorf("failed to save aigogo.json: %w", err)
	}

	fmt.Printf("✓ Added %s %s to %s dependencies\n", pkgName, version, depType)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'aigogo validate' to check your dependencies")
	fmt.Println("  2. Run 'aigogo scan' to detect any missing dependencies")

	return nil
}

// suggestVersionFormat provides a language-specific version format suggestion
func suggestVersionFormat(language, pkgName string) string {
	switch strings.ToLower(language) {
	case "python":
		return "(e.g., >=2.31.0,<3.0.0)"
	case "javascript":
		return "(e.g., ^1.6.0 or ~1.6.0)"
	case "go":
		return "(e.g., v1.2.3)"
	case "rust":
		return "(e.g., 1.0 or ^1.0)"
	default:
		return ""
	}
}

// addDependenciesFromPyproject imports dependencies from pyproject.toml
func addDependenciesFromPyproject(isDev bool) error {
	// Find and load manifest (supports subdirectories)
	m, manifestDir, err := manifest.FindManifest()
	if err != nil {
		return fmt.Errorf("failed to find aigogo.json: %w\nRun 'aigogo init' first", err)
	}

	manifestPath := filepath.Join(manifestDir, "aigogo.json")

	// Check that this is a Python project
	if strings.ToLower(m.Language.Name) != "python" {
		return fmt.Errorf("--from-pyproject is only supported for Python projects (current language: %s)", m.Language.Name)
	}

	// Find pyproject.toml
	pyprojectPath, err := pyproject.FindPyProject(manifestDir)
	if err != nil {
		return fmt.Errorf("failed to find pyproject.toml: %w", err)
	}

	fmt.Printf("📦 Reading dependencies from: %s\n", pyprojectPath)

	// Parse pyproject.toml
	pyprojData, err := pyproject.Parse(pyprojectPath)
	if err != nil {
		return err
	}

	// Extract dependencies
	deps, err := pyproject.ExtractDependencies(pyprojData)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Detected format: %s\n\n", deps.Format)

	// Initialize dependencies if nil
	if m.Dependencies == nil {
		m.Dependencies = &manifest.Dependencies{
			Runtime: []manifest.Dependency{},
			Dev:     []manifest.Dependency{},
		}
	}

	// Update Python version if found
	if deps.PythonVersion != "" && m.Language.Version == "" {
		m.Language.Version = deps.PythonVersion
		fmt.Printf("✓ Set Python version requirement: %s\n", deps.PythonVersion)
	}

	var added int
	var skipped int

	if isDev {
		// Add only dev dependencies
		fmt.Printf("Adding %d development dependencies...\n\n", len(deps.Dev))
		for _, dep := range deps.Dev {
			// Check if already exists
			exists := false
			for _, existing := range m.Dependencies.Dev {
				if existing.Package == dep.Package {
					fmt.Printf("⚠ Skipping '%s' (already exists)\n", dep.Package)
					exists = true
					skipped++
					break
				}
			}
			if !exists {
				m.Dependencies.Dev = append(m.Dependencies.Dev, dep)
				fmt.Printf("✓ Added %s %s\n", dep.Package, dep.Version)
				added++
			}
		}
	} else {
		// Add only runtime dependencies
		fmt.Printf("Adding %d runtime dependencies...\n\n", len(deps.Runtime))
		for _, dep := range deps.Runtime {
			// Check if already exists
			exists := false
			for _, existing := range m.Dependencies.Runtime {
				if existing.Package == dep.Package {
					fmt.Printf("⚠ Skipping '%s' (already exists)\n", dep.Package)
					exists = true
					skipped++
					break
				}
			}
			if !exists {
				m.Dependencies.Runtime = append(m.Dependencies.Runtime, dep)
				fmt.Printf("✓ Added %s %s\n", dep.Package, dep.Version)
				added++
			}
		}
	}

	// Save updated manifest
	if err := manifest.Save(manifestPath, m); err != nil {
		return fmt.Errorf("failed to save aigogo.json: %w", err)
	}

	fmt.Printf("\n✓ Successfully added %d dependencies", added)
	if skipped > 0 {
		fmt.Printf(" (%d skipped)", skipped)
	}
	fmt.Println()

	if deps.PythonVersion != "" && m.Language.Version != deps.PythonVersion {
		fmt.Printf("\n💡 Note: Your aigogo.json Python version (%s) differs from pyproject.toml (%s)\n",
			m.Language.Version, deps.PythonVersion)
	}

	return nil
}
