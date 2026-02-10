package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aupeachmo/aigogo/pkg/docker"
	"github.com/aupeachmo/aigogo/pkg/imports"
	"github.com/aupeachmo/aigogo/pkg/lockfile"
	"github.com/aupeachmo/aigogo/pkg/store"
)

func installCmd() *Command {
	return &Command{
		Name:        "install",
		Description: "Install packages from aigogo.lock",
		Run: func(args []string) error {
			return runInstall()
		},
	}
}

func runInstall() error {
	// Find lock file
	lockPath, lock, err := lockfile.FindLockFile()
	if err != nil {
		return fmt.Errorf("failed to find aigogo.lock: %w\nRun 'aigg add <package>' first to add packages", err)
	}

	projectDir := filepath.Dir(lockPath)
	fmt.Printf("Installing packages from %s\n\n", lockPath)

	if len(lock.Packages) == 0 {
		fmt.Println("No packages to install")
		return nil
	}

	// Initialize store
	cas, err := store.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	// Initialize setup manager
	setupMgr, err := imports.NewSetupManager(projectDir)
	if err != nil {
		return fmt.Errorf("failed to initialize imports manager: %w", err)
	}

	// Clean existing imports
	if err := setupMgr.Clean(); err != nil {
		return fmt.Errorf("failed to clean existing imports: %w", err)
	}

	// Track languages for namespace setup
	hasPython := false
	hasJavaScript := false

	// Count packages by language
	for _, pkg := range lock.Packages {
		switch pkg.Language {
		case "python":
			hasPython = true
		case "javascript", "typescript":
			hasJavaScript = true
		}
	}

	// Setup namespaces
	if hasPython {
		if err := setupMgr.SetupPythonNamespace(); err != nil {
			return fmt.Errorf("failed to setup Python namespace: %w", err)
		}
	}
	if hasJavaScript {
		if err := setupMgr.SetupJavaScriptNamespace(); err != nil {
			return fmt.Errorf("failed to setup JavaScript namespace: %w", err)
		}
	}

	// Install each package
	var installed, fetched int
	for name, pkg := range lock.Packages {
		hash := pkg.GetIntegrityHash()

		// Check if in store
		if !cas.Has(hash) {
			// Need to fetch from source
			fmt.Printf("Fetching %s from %s...\n", name, pkg.Source)
			if err := fetchAndStore(cas, pkg); err != nil {
				return fmt.Errorf("failed to fetch %s: %w", name, err)
			}

			// Verify hash matches
			if !cas.Has(hash) {
				return fmt.Errorf("integrity check failed for %s: hash mismatch", name)
			}
			fetched++
		}

		// Get stored package
		storedPkg, err := cas.Get(hash)
		if err != nil {
			return fmt.Errorf("failed to get package %s from store: %w", name, err)
		}

		// Create symlink
		storePath := cas.GetPath(hash)
		if err := setupMgr.CreatePackageLink(name, pkg.Language, storePath); err != nil {
			return fmt.Errorf("failed to create link for %s: %w", name, err)
		}

		// Display info
		linkName := name
		if pkg.Language == "python" {
			linkName = lockfile.NormalizeName(name)
		}

		fmt.Printf("✓ Installed %s (%d files)\n", name, len(pkg.Files))
		_ = storedPkg // Use variable
		installed++

		// Show import hint
		switch pkg.Language {
		case "python":
			fmt.Printf("  import: from aigogo.%s import ...\n", linkName)
		case "javascript", "typescript":
			fmt.Printf("  import: import ... from '@aigogo/%s'\n", name)
		}
	}

	// Update .gitignore
	if err := setupMgr.UpdateGitignore(); err != nil {
		fmt.Printf("⚠ Warning: failed to update .gitignore: %v\n", err)
	}

	// Auto-configure Python path via .pth file
	// Note: Clean() already removed any stale .pth file at the start of install.
	pythonPathConfigured := false
	if hasPython {
		if err := imports.InstallPthFile(setupMgr.GetImportsDir()); err != nil {
			fmt.Printf("⚠ Could not auto-configure Python path: %v\n", err)
		} else {
			pythonPathConfigured = true
		}
	}

	fmt.Printf("\n✓ Installed %d package(s)", installed)
	if fetched > 0 {
		fmt.Printf(" (%d fetched)", fetched)
	}
	fmt.Println()

	// Generate Node.js register script
	// Note: Clean() already removed any stale register script at the start of install.
	jsRegisterInstalled := false
	if hasJavaScript {
		if err := imports.InstallRegisterScript(projectDir); err != nil {
			fmt.Printf("⚠ Warning: failed to create register script: %v\n", err)
		} else {
			jsRegisterInstalled = true
		}
	}

	// Print setup hints
	fmt.Println("\nTo use installed packages:")
	if hasPython {
		if pythonPathConfigured {
			fmt.Println("  Python: Path auto-configured via .pth file")
			fmt.Println("    from aigogo.<package_name> import ...")
		} else {
			fmt.Println("  Python: Add to PYTHONPATH (auto-configuration failed):")
			fmt.Printf("    export PYTHONPATH=\"%s:$PYTHONPATH\"\n", setupMgr.GetImportsDir())
		}
	}
	if hasJavaScript {
		if jsRegisterInstalled {
			fmt.Println("  JavaScript: Add to entry point (CommonJS):")
			fmt.Println("    require('./.aigogo/register');")
			fmt.Println("  Or use as preload (CommonJS and ESM):")
			fmt.Println("    node --require ./.aigogo/register.js app.js")
		} else {
			fmt.Println("  JavaScript: Add to NODE_PATH:")
			fmt.Printf("    export NODE_PATH=\"%s:$NODE_PATH\"\n", setupMgr.GetImportsDir())
		}
	}

	return nil
}

// fetchAndStore pulls a package from the registry and stores it in the CAS
func fetchAndStore(cas *store.Store, pkg lockfile.LockedPackage) error {
	// Pull the package using existing Puller
	puller := docker.NewPuller()
	if err := puller.Pull(pkg.Source); err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}

	// Extract to temp directory
	tmpDir, err := os.MkdirTemp("", "aigogo-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	extractor := docker.NewExtractor()
	extractedFiles, err := extractor.Extract(pkg.Source, tmpDir, true)
	if err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Convert to relative paths
	var relFiles []string
	for _, f := range extractedFiles {
		relPath, err := filepath.Rel(tmpDir, f)
		if err != nil {
			return err
		}
		relFiles = append(relFiles, relPath)
	}

	// Read manifest if present
	manifestPath := filepath.Join(tmpDir, "aigogo.json")
	var manifestData []byte
	if _, err := os.Stat(manifestPath); err == nil {
		manifestData, _ = os.ReadFile(manifestPath)
	} else {
		// Create minimal manifest
		manifest := map[string]interface{}{
			"name":     lockfile.GetPackageName(pkg.Source),
			"version":  pkg.Version,
			"language": map[string]string{"name": pkg.Language},
		}
		manifestData, _ = json.MarshalIndent(manifest, "", "  ")
	}

	// Store in CAS
	hash, err := cas.Store(tmpDir, relFiles, manifestData)
	if err != nil {
		return fmt.Errorf("failed to store: %w", err)
	}

	// Verify integrity
	expectedHash := pkg.GetIntegrityHash()
	if hash != expectedHash {
		// Clean up
		_ = cas.Delete(hash)
		return fmt.Errorf("integrity mismatch: expected %s, got %s", expectedHash, hash)
	}

	// Make read-only
	if err := cas.MakeReadOnly(hash); err != nil {
		fmt.Printf("⚠ Warning: failed to make files read-only: %v\n", err)
	}

	return nil
}
