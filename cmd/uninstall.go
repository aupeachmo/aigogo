package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aupeachmo/aigogo/pkg/imports"
)

func uninstallCmd() *Command {
	return &Command{
		Name:        "uninstall",
		Description: "Remove all installed packages and import configuration from this project",
		Run: func(args []string) error {
			return runUninstall()
		},
	}
}

func runUninstall() error {
	// Find project directory by looking for .aigogo/ or aigogo.lock
	projectDir, err := findProjectDir()
	if err != nil {
		return fmt.Errorf("not an aigogo project (no .aigogo/ directory or aigogo.lock found)")
	}

	aigogoDir := filepath.Join(projectDir, imports.ImportsDir)

	// Check if there's anything to uninstall
	if _, err := os.Stat(aigogoDir); os.IsNotExist(err) {
		fmt.Println("Nothing to uninstall (.aigogo/ directory not found)")
		return nil
	}

	// Remove .pth file from Python site-packages (only if tracking file exists)
	pthTrackingPath := filepath.Join(aigogoDir, ".pth-location")
	if _, err := os.Stat(pthTrackingPath); err == nil {
		if err := imports.RemovePthFile(projectDir); err != nil {
			fmt.Printf("⚠ Warning: failed to remove Python .pth file: %v\n", err)
		} else {
			fmt.Println("✓ Removed Python .pth file from site-packages")
		}
	}

	// Remove register.js
	registerPath := filepath.Join(aigogoDir, "register.js")
	if _, err := os.Stat(registerPath); err == nil {
		if err := imports.RemoveRegisterScript(projectDir); err != nil {
			fmt.Printf("⚠ Warning: failed to remove Node.js register script: %v\n", err)
		} else {
			fmt.Println("✓ Removed Node.js register script")
		}
	}

	// Remove the entire .aigogo/ directory
	if err := os.RemoveAll(aigogoDir); err != nil {
		return fmt.Errorf("failed to remove .aigogo/ directory: %w", err)
	}
	fmt.Println("✓ Removed .aigogo/ directory")

	fmt.Println("\nUninstall complete. The aigogo.lock file has been preserved.")
	fmt.Println("Run 'aigogo install' to reinstall packages.")

	return nil
}

// findProjectDir looks for the project root by searching for .aigogo/ or aigogo.lock
// in the current directory and parent directories.
func findProjectDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check for .aigogo/ directory
		if _, err := os.Stat(filepath.Join(dir, imports.ImportsDir)); err == nil {
			return dir, nil
		}
		// Check for aigogo.lock
		if _, err := os.Stat(filepath.Join(dir, "aigogo.lock")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not found")
		}
		dir = parent
	}
}
