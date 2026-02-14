package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aupeachmo/aigogo/pkg/docker"
	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func diffCmd() *Command {
	flags := flag.NewFlagSet("diff", flag.ContinueOnError)
	remote := flags.Bool("remote", false, "Compare against a remote registry image")
	summary := flags.Bool("summary", false, "Show compact M/A/D summary only")

	return &Command{
		Name:        "diff",
		Description: "Compare package versions",
		Flags:       flags,
		Run: func(args []string) error {
			differ := docker.NewDiffer()

			switch {
			case len(args) == 0 && !*remote:
				// Working dir vs latest local build
				return diffWorkingDirVsLocal(differ, *summary)

			case len(args) == 1 && !*remote:
				// Working dir vs specified local build
				return diffWorkingDirVsRef(differ, args[0], *summary)

			case len(args) == 2 && !*remote:
				// Local build vs local build
				return diffLocalVsLocal(differ, args[0], args[1], *summary)

			case len(args) == 1 && *remote:
				// Latest local build vs remote ref
				return diffLocalVsRemote(differ, args[0], *summary)

			case len(args) == 2 && *remote:
				// Local build (arg1) vs remote (arg2)
				return diffLocalRefVsRemote(differ, args[0], args[1], *summary)

			default:
				return fmt.Errorf("usage: aigg diff [--remote] [--summary] [<ref-a>] [<ref-b>]\n\nExamples:\n  aigg diff                     # Working dir vs latest local build\n  aigg diff utils:1.0.0         # Working dir vs specified build\n  aigg diff utils:1.0.0 utils:1.1.0  # Compare two local builds\n  aigg diff --remote docker.io/org/utils:1.0.0  # Local vs remote\n  aigg diff --remote utils:1.0.0 docker.io/org/utils:1.0.0  # Specific local vs remote")
			}
		},
	}
}

// diffWorkingDirVsLocal compares the current working directory against the latest local build
func diffWorkingDirVsLocal(differ *docker.Differ, summaryOnly bool) error {
	m, manifestDir, err := manifest.FindManifest()
	if err != nil {
		return fmt.Errorf("failed to find manifest: %w", err)
	}

	if m.Name == "" || m.Version == "" {
		return fmt.Errorf("aigogo.json must have name and version fields")
	}

	ref := fmt.Sprintf("%s:%s", m.Name, m.Version)
	return diffWorkingDirVsRef(differ, ref, summaryOnly, withManifest(m, manifestDir))
}

type diffOpts struct {
	m           *manifest.Manifest
	manifestDir string
}

// diffWorkingDirVsRef compares the current working directory against a specified local build
func diffWorkingDirVsRef(differ *docker.Differ, ref string, summaryOnly bool, opts ...diffOpts) error {
	var m *manifest.Manifest
	var manifestDir string

	if len(opts) > 0 {
		m = opts[0].m
		manifestDir = opts[0].manifestDir
	} else {
		var err error
		m, manifestDir, err = manifest.FindManifest()
		if err != nil {
			return fmt.Errorf("failed to find manifest: %w", err)
		}
	}

	// Check that the local build exists
	if !docker.ImageExistsInCache(ref) {
		return fmt.Errorf("local build not found: %s\nBuild it first with: aigg build %s", ref, ref)
	}

	// Copy working dir files to a temp dir (based on manifest discovery)
	workDir, err := copyWorkingDirFiles(m, manifestDir)
	if err != nil {
		return fmt.Errorf("failed to collect working directory files: %w", err)
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	// Extract the local build to a temp dir
	buildDir, err := differ.ExtractToTemp(ref)
	if err != nil {
		return fmt.Errorf("failed to extract local build: %w", err)
	}
	defer func() { _ = os.RemoveAll(buildDir) }()

	result, err := differ.CompareDirs(buildDir, workDir)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	fmt.Print(docker.FormatDiff(result, summaryOnly))
	return nil
}

// diffLocalVsLocal compares two local builds
func diffLocalVsLocal(differ *docker.Differ, refA, refB string, summaryOnly bool) error {
	if !docker.ImageExistsInCache(refA) {
		return fmt.Errorf("local build not found: %s", refA)
	}
	if !docker.ImageExistsInCache(refB) {
		return fmt.Errorf("local build not found: %s", refB)
	}

	dirA, err := differ.ExtractToTemp(refA)
	if err != nil {
		return fmt.Errorf("failed to extract %s: %w", refA, err)
	}
	defer func() { _ = os.RemoveAll(dirA) }()

	dirB, err := differ.ExtractToTemp(refB)
	if err != nil {
		return fmt.Errorf("failed to extract %s: %w", refB, err)
	}
	defer func() { _ = os.RemoveAll(dirB) }()

	result, err := differ.CompareDirs(dirA, dirB)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	fmt.Print(docker.FormatDiff(result, summaryOnly))
	return nil
}

// diffLocalVsRemote compares the latest local build against a remote ref
func diffLocalVsRemote(differ *docker.Differ, remoteRef string, summaryOnly bool) error {
	m, _, err := manifest.FindManifest()
	if err != nil {
		return fmt.Errorf("failed to find manifest: %w", err)
	}

	if m.Name == "" || m.Version == "" {
		return fmt.Errorf("aigogo.json must have name and version fields")
	}

	localRef := fmt.Sprintf("%s:%s", m.Name, m.Version)
	return diffLocalRefVsRemote(differ, localRef, remoteRef, summaryOnly)
}

// diffLocalRefVsRemote compares a specific local build against a remote ref
func diffLocalRefVsRemote(differ *docker.Differ, localRef, remoteRef string, summaryOnly bool) error {
	if !docker.ImageExistsInCache(localRef) {
		return fmt.Errorf("local build not found: %s", localRef)
	}

	fmt.Printf("Comparing local %s vs remote %s...\n", localRef, remoteRef)

	result, err := differ.CompareWithRemote(localRef, remoteRef)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	fmt.Print(docker.FormatDiff(result, summaryOnly))
	return nil
}

// copyWorkingDirFiles discovers files from the manifest and copies them to a temp directory
func copyWorkingDirFiles(m *manifest.Manifest, manifestDir string) (string, error) {
	// Save and restore working directory
	origDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if err := os.Chdir(manifestDir); err != nil {
		return "", err
	}
	defer func() { _ = os.Chdir(origDir) }()

	discovery, err := manifest.NewFileDiscovery(".", m.Files.Exclude)
	if err != nil {
		return "", fmt.Errorf("failed to initialize file discovery: %w", err)
	}

	files, err := discovery.Discover(m.Files, m.Language)
	if err != nil {
		return "", fmt.Errorf("failed to discover files: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "aigogo-diff-workdir-*")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		srcPath := filepath.Join(manifestDir, file)
		dstPath := filepath.Join(tmpDir, file)

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", err
		}

		content, err := os.ReadFile(srcPath)
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to read %s: %w", file, err)
		}

		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", err
		}
	}

	return tmpDir, nil
}

func withManifest(m *manifest.Manifest, dir string) diffOpts {
	return diffOpts{m: m, manifestDir: dir}
}
