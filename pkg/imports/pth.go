package imports

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// pthFileName is the name of the .pth file written to site-packages
	pthFileName = "aigogo.pth"
	// pthLocationFile tracks where the .pth file was written
	pthLocationFile = ".pth-location"
)

// InstallPthFile writes an aigogo.pth file into the active Python environment's
// site-packages directory. This is the standard mechanism (used by pip install -e)
// for adding directories to Python's import path.
func InstallPthFile(importsDir string) error {
	sitePackages, err := findSitePackages()
	if err != nil {
		return err
	}

	absImportsDir, err := filepath.Abs(importsDir)
	if err != nil {
		return fmt.Errorf("failed to resolve imports directory: %w", err)
	}

	pthPath := filepath.Join(sitePackages, pthFileName)
	if err := os.WriteFile(pthPath, []byte(absImportsDir+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", pthPath, err)
	}

	// Record the .pth file location for cleanup
	aigogoDir := filepath.Dir(importsDir) // .aigogo/
	trackingPath := filepath.Join(aigogoDir, pthLocationFile)
	if err := os.WriteFile(trackingPath, []byte(pthPath+"\n"), 0644); err != nil {
		// Non-fatal: .pth file was written successfully
		return fmt.Errorf("warning: .pth file installed but failed to record location: %w", err)
	}

	return nil
}

// RemovePthFile reads the .pth-location tracking file and removes the .pth file
// if it exists. projectDir is the project root (where .aigogo/ lives).
func RemovePthFile(projectDir string) error {
	trackingPath := filepath.Join(projectDir, ImportsDir, pthLocationFile)

	data, err := os.ReadFile(trackingPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No tracking file, nothing to clean up
		}
		return fmt.Errorf("failed to read .pth-location: %w", err)
	}

	pthPath := strings.TrimSpace(string(data))
	if pthPath == "" {
		return nil
	}

	if err := os.Remove(pthPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", pthPath, err)
	}

	// Remove tracking file
	if err := os.Remove(trackingPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove tracking file: %w", err)
	}

	return nil
}

// findSitePackages locates the site-packages directory for the active Python environment.
func findSitePackages() (string, error) {
	// Check $VIRTUAL_ENV first
	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		return findVenvSitePackages(venv)
	}

	// Fall back to python3 sysconfig
	return findSystemSitePackages()
}

// findVenvSitePackages finds site-packages inside a virtualenv by globbing for
// the pythonX.Y directory.
func findVenvSitePackages(venvPath string) (string, error) {
	pattern := filepath.Join(venvPath, "lib", "python*", "site-packages")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to glob virtualenv site-packages: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no site-packages found in virtualenv %s", venvPath)
	}
	// Use the first match (there should only be one)
	sp := matches[0]
	if _, err := os.Stat(sp); err != nil {
		return "", fmt.Errorf("site-packages directory does not exist: %s", sp)
	}
	return sp, nil
}

// findSystemSitePackages uses python3 sysconfig to find the system site-packages.
func findSystemSitePackages() (string, error) {
	cmd := exec.Command("python3", "-c", "import sysconfig; print(sysconfig.get_path('purelib'))")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("python3 not found or failed: %w (install python3 or activate a virtualenv)", err)
	}

	sp := strings.TrimSpace(string(out))
	if sp == "" {
		return "", fmt.Errorf("python3 returned empty site-packages path")
	}

	if _, err := os.Stat(sp); err != nil {
		return "", fmt.Errorf("site-packages directory does not exist: %s", sp)
	}

	return sp, nil
}
