package imports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallPthFileWithVirtualEnv(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake virtualenv with site-packages
	sitePackages := filepath.Join(tmpDir, "venv", "lib", "python3.11", "site-packages")
	if err := os.MkdirAll(sitePackages, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a fake project with .aigogo/imports
	projectDir := filepath.Join(tmpDir, "project")
	importsDir := filepath.Join(projectDir, ".aigogo", "imports")
	if err := os.MkdirAll(importsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Set VIRTUAL_ENV
	t.Setenv("VIRTUAL_ENV", filepath.Join(tmpDir, "venv"))

	if err := InstallPthFile(importsDir); err != nil {
		t.Fatalf("InstallPthFile failed: %v", err)
	}

	// Verify .pth file was created
	pthPath := filepath.Join(sitePackages, pthFileName)
	content, err := os.ReadFile(pthPath)
	if err != nil {
		t.Fatalf("Failed to read .pth file: %v", err)
	}

	absImportsDir, _ := filepath.Abs(importsDir)
	if strings.TrimSpace(string(content)) != absImportsDir {
		t.Errorf(".pth content = %q, want %q", strings.TrimSpace(string(content)), absImportsDir)
	}

	// Verify tracking file was created
	trackingPath := filepath.Join(projectDir, ".aigogo", pthLocationFile)
	trackingContent, err := os.ReadFile(trackingPath)
	if err != nil {
		t.Fatalf("Failed to read tracking file: %v", err)
	}

	if strings.TrimSpace(string(trackingContent)) != pthPath {
		t.Errorf("tracking file content = %q, want %q", strings.TrimSpace(string(trackingContent)), pthPath)
	}
}

func TestInstallPthFileNoVenvNoSitePackages(t *testing.T) {
	tmpDir := t.TempDir()

	// Set VIRTUAL_ENV to a path without site-packages
	t.Setenv("VIRTUAL_ENV", filepath.Join(tmpDir, "nonexistent-venv"))

	importsDir := filepath.Join(tmpDir, "project", ".aigogo", "imports")
	if err := os.MkdirAll(importsDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := InstallPthFile(importsDir)
	if err == nil {
		t.Error("Expected error when no site-packages exists")
	}
}

func TestRemovePthFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake .pth file
	sitePackages := filepath.Join(tmpDir, "site-packages")
	if err := os.MkdirAll(sitePackages, 0755); err != nil {
		t.Fatal(err)
	}
	pthPath := filepath.Join(sitePackages, pthFileName)
	if err := os.WriteFile(pthPath, []byte("/some/path\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create tracking file
	projectDir := filepath.Join(tmpDir, "project")
	aigogoDir := filepath.Join(projectDir, ImportsDir)
	if err := os.MkdirAll(aigogoDir, 0755); err != nil {
		t.Fatal(err)
	}
	trackingPath := filepath.Join(aigogoDir, pthLocationFile)
	if err := os.WriteFile(trackingPath, []byte(pthPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Remove
	if err := RemovePthFile(projectDir); err != nil {
		t.Fatalf("RemovePthFile failed: %v", err)
	}

	// Verify .pth file removed
	if _, err := os.Stat(pthPath); !os.IsNotExist(err) {
		t.Error(".pth file should have been removed")
	}

	// Verify tracking file removed
	if _, err := os.Stat(trackingPath); !os.IsNotExist(err) {
		t.Error("tracking file should have been removed")
	}
}

func TestRemovePthFileNoTrackingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// No tracking file — should not error
	if err := RemovePthFile(tmpDir); err != nil {
		t.Errorf("RemovePthFile should not error when no tracking file exists: %v", err)
	}
}

func TestRemovePthFilePthAlreadyGone(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tracking file that points to a non-existent .pth
	projectDir := filepath.Join(tmpDir, "project")
	aigogoDir := filepath.Join(projectDir, ImportsDir)
	if err := os.MkdirAll(aigogoDir, 0755); err != nil {
		t.Fatal(err)
	}
	trackingPath := filepath.Join(aigogoDir, pthLocationFile)
	if err := os.WriteFile(trackingPath, []byte("/nonexistent/aigogo.pth\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should not error even though .pth doesn't exist
	if err := RemovePthFile(projectDir); err != nil {
		t.Errorf("RemovePthFile should not error when .pth is already gone: %v", err)
	}

	// Tracking file should still be cleaned up
	if _, err := os.Stat(trackingPath); !os.IsNotExist(err) {
		t.Error("tracking file should have been removed")
	}
}

func TestInstallAndRemovePthFileRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake virtualenv with site-packages
	sitePackages := filepath.Join(tmpDir, "venv", "lib", "python3.12", "site-packages")
	if err := os.MkdirAll(sitePackages, 0755); err != nil {
		t.Fatal(err)
	}

	// Create project structure
	projectDir := filepath.Join(tmpDir, "project")
	importsDir := filepath.Join(projectDir, ".aigogo", "imports")
	if err := os.MkdirAll(importsDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("VIRTUAL_ENV", filepath.Join(tmpDir, "venv"))

	// Install
	if err := InstallPthFile(importsDir); err != nil {
		t.Fatalf("InstallPthFile failed: %v", err)
	}

	// Verify .pth exists
	pthPath := filepath.Join(sitePackages, pthFileName)
	if _, err := os.Stat(pthPath); os.IsNotExist(err) {
		t.Fatal(".pth file should exist after install")
	}

	// Remove
	if err := RemovePthFile(projectDir); err != nil {
		t.Fatalf("RemovePthFile failed: %v", err)
	}

	// Verify .pth removed
	if _, err := os.Stat(pthPath); !os.IsNotExist(err) {
		t.Error(".pth file should be removed after cleanup")
	}
}

func TestFindVenvSitePackages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple python version dirs (edge case)
	sp := filepath.Join(tmpDir, "lib", "python3.11", "site-packages")
	if err := os.MkdirAll(sp, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := findVenvSitePackages(tmpDir)
	if err != nil {
		t.Fatalf("findVenvSitePackages failed: %v", err)
	}

	if result != sp {
		t.Errorf("result = %q, want %q", result, sp)
	}
}

func TestFindVenvSitePackagesNoMatch(t *testing.T) {
	tmpDir := t.TempDir()

	// No lib directory at all
	_, err := findVenvSitePackages(tmpDir)
	if err == nil {
		t.Error("Expected error when no site-packages found")
	}
}
