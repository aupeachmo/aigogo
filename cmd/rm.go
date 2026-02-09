package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func rmCmd() *Command {
	return &Command{
		Name:        "rm",
		Description: "Remove files or dependencies from aigogo.json",
		Run: func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: aigogo rm <file|dep|dev> [args...]\n\nSubcommands:\n  file <path>...  Remove files from include list\n  dep <pkg>       Remove runtime dependency\n  dev <pkg>       Remove development dependency")
			}

			subcommand := args[0]
			subArgs := args[1:]

			switch subcommand {
			case "file":
				return rmFiles(subArgs)
			case "dep":
				return rmDependency(subArgs, false)
			case "dev":
				return rmDependency(subArgs, true)
			default:
				return fmt.Errorf("unknown subcommand '%s'\nValid subcommands: file, dep, dev", subcommand)
			}
		},
	}
}

func rmFiles(args []string) error {
	// Find and load manifest (supports subdirectories)
	m, manifestDir, err := manifest.FindManifest()
	if err != nil {
		return fmt.Errorf("failed to find aigogo.json: %w\nRun 'aigogo init' first", err)
	}

	manifestPath := filepath.Join(manifestDir, "aigogo.json")

	// Check if files.include is "auto"
	if str, ok := m.Files.Include.(string); ok && str == "auto" {
		return fmt.Errorf("files.include is set to 'auto'\nCannot remove individual files when using auto-discovery")
	}

	// Get existing patterns
	existingPatterns, _ := m.Files.GetIncludePatterns()
	if len(existingPatterns) == 0 {
		return fmt.Errorf("no files in include list")
	}

	reader := bufio.NewReader(os.Stdin)

	// Get file paths
	var filePaths []string
	if len(args) > 0 {
		filePaths = args
	} else {
		// Show available files
		fmt.Println("Current files in include list:")
		for i, path := range existingPatterns {
			fmt.Printf("  %d. %s\n", i+1, path)
		}
		fmt.Println()

		fmt.Print("File path(s) to remove (space-separated): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read file paths: %w", err)
		}
		filePaths = strings.Fields(strings.TrimSpace(input))
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("at least one file path is required")
	}

	// Remove specified files
	var removedFiles []string
	newPatterns := []string{}

	for _, existing := range existingPatterns {
		shouldRemove := false
		for _, toRemove := range filePaths {
			if existing == toRemove {
				shouldRemove = true
				removedFiles = append(removedFiles, existing)
				break
			}
		}
		if !shouldRemove {
			newPatterns = append(newPatterns, existing)
		}
	}

	if len(removedFiles) == 0 {
		return fmt.Errorf("no matching files found in include list")
	}

	// Update include list (or set to empty array)
	if len(newPatterns) == 0 {
		m.Files.Include = []string{}
	} else {
		m.Files.Include = newPatterns
	}

	// Save updated manifest
	if err := manifest.Save(manifestPath, m); err != nil {
		return fmt.Errorf("failed to save aigogo.json: %w", err)
	}

	fmt.Printf("✓ Removed %d file(s) from include list:\n", len(removedFiles))
	for _, f := range removedFiles {
		fmt.Printf("  - %s\n", f)
	}

	// Show remaining files
	if len(newPatterns) > 0 {
		fmt.Printf("\nRemaining files (%d):\n", len(newPatterns))
		for _, f := range newPatterns {
			fmt.Printf("  - %s\n", f)
		}
	} else {
		fmt.Println("\nNo files remaining in include list")
	}

	return nil
}

func rmDependency(args []string, isDev bool) error {
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

	// Check if dependencies exist
	if m.Dependencies == nil {
		return fmt.Errorf("no dependencies found in aigogo.json")
	}

	targetList := m.Dependencies.Runtime
	if isDev {
		targetList = m.Dependencies.Dev
	}

	if len(targetList) == 0 {
		return fmt.Errorf("no %s dependencies found in aigogo.json", depType)
	}

	reader := bufio.NewReader(os.Stdin)

	// Get package name
	var pkgName string
	if len(args) > 0 {
		pkgName = args[0]
	} else {
		// Show available packages
		fmt.Printf("Current %s dependencies:\n", depType)
		for i, dep := range targetList {
			fmt.Printf("  %d. %s (%s)\n", i+1, dep.Package, dep.Version)
		}
		fmt.Println()

		fmt.Print("Package name to remove: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read package name: %w", err)
		}
		pkgName = strings.TrimSpace(input)
	}

	if pkgName == "" {
		return fmt.Errorf("package name is required")
	}

	// Find and remove the dependency
	found := false
	var removedVersion string
	newList := make([]manifest.Dependency, 0)

	for _, dep := range targetList {
		if dep.Package == pkgName {
			found = true
			removedVersion = dep.Version
			// Skip this one (don't add to newList)
		} else {
			newList = append(newList, dep)
		}
	}

	if !found {
		return fmt.Errorf("package '%s' not found in %s dependencies", pkgName, depType)
	}

	// Update the dependencies
	if isDev {
		m.Dependencies.Dev = newList
	} else {
		m.Dependencies.Runtime = newList
	}

	// Clean up if no dependencies left
	if len(m.Dependencies.Runtime) == 0 && len(m.Dependencies.Dev) == 0 {
		m.Dependencies = nil
	}

	// Save updated manifest
	if err := manifest.Save(manifestPath, m); err != nil {
		return fmt.Errorf("failed to save aigogo.json: %w", err)
	}

	fmt.Printf("✓ Removed %s (%s) from %s dependencies\n", pkgName, removedVersion, depType)

	// Show remaining dependencies
	if m.Dependencies != nil {
		remaining := m.Dependencies.Runtime
		if isDev {
			remaining = m.Dependencies.Dev
		}

		if len(remaining) > 0 {
			fmt.Printf("\nRemaining %s dependencies (%d):\n", depType, len(remaining))
			for _, dep := range remaining {
				fmt.Printf("  - %s (%s)\n", dep.Package, dep.Version)
			}
		} else {
			fmt.Printf("\nNo %s dependencies remaining\n", depType)
		}
	}

	return nil
}
