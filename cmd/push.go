package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/aupeachmo/aigogo/pkg/docker"
)

func pushCmd() *Command {
	flags := flag.NewFlagSet("push", flag.ExitOnError)
	from := flags.String("from", "", "Push from existing local build (required)")
	dryRun := flags.Bool("dry-run", false, "Check if push is needed without pushing")

	return &Command{
		Name:        "push",
		Description: "Push an agent to a registry",
		Flags:       flags,
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigg push <registry>/<name>:<tag> --from <local-build>")
			}

			imageRef := args[0]

			// Require --from flag
			if *from == "" {
				return fmt.Errorf("--from flag is required\n\nWorkflow:\n  1. aigg build <name>:<tag>\n  2. aigg push %s --from <name>:<tag>\n\nExample:\n  aigg build utils:1.0.0\n  aigg push %s --from utils:1.0.0", imageRef, imageRef)
			}

			if *dryRun {
				return pushDryRun(imageRef, *from)
			}

			// Push from the specified local build
			return pushFromLocalBuild(imageRef, *from)
		},
	}
}

// pushDryRun checks if a push is needed without actually pushing.
// Uses CompareWithRemote which handles Tier 1 (digest) → Tier 2 (content) fallback.
func pushDryRun(registryRef, localRef string) error {
	if !docker.ImageExistsInCache(localRef) {
		return fmt.Errorf("local build not found: %s\nBuild it first with: aigg build %s", localRef, localRef)
	}

	fmt.Printf("Checking if push is needed: %s -> %s\n", localRef, registryRef)

	differ := docker.NewDiffer()

	result, err := differ.CompareWithRemote(localRef, registryRef)
	if err != nil {
		// Remote doesn't exist or is unreachable — push is needed
		fmt.Println("Remote image not found or unreachable — push needed.")
		return nil
	}

	if result.Identical {
		fmt.Println("Remote is already up-to-date — no push needed.")
		return nil
	}

	fmt.Println("Changes detected — push needed.")
	fmt.Println()
	fmt.Print(docker.FormatSummary(result))
	return nil
}

// pushFromLocalBuild pushes an existing local build to a registry
func pushFromLocalBuild(registryRef, localRef string) error {
	// Check if local build exists
	if !docker.ImageExistsInCache(localRef) {
		return fmt.Errorf("local build not found: %s\nBuild it first with: aigg build %s", localRef, localRef)
	}

	fmt.Printf("Pushing local build %s to %s...\n", localRef, registryRef)

	// Get cache directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	cacheDir := home + "/.aigogo/cache"

	// Sanitize local ref to get cache path
	sanitized := docker.SanitizeImageRef(localRef)
	localPath := cacheDir + "/" + sanitized

	// Read files from local cache
	files, err := getFilesFromLocalBuild(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local build: %w", err)
	}

	fmt.Printf("  Found %d file(s) in local build\n", len(files))

	// Build Docker image from local files
	fmt.Println("Building image for registry...")
	builder := docker.NewBuilder()

	// Create a simple manifest for the builder
	simpleManifest := map[string]interface{}{
		"name":    localRef,
		"version": "local",
	}

	if err := builder.BuildImageFromPath(registryRef, localPath, files, simpleManifest); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	// Push to registry
	fmt.Printf("Pushing to %s...\n", registryRef)
	pusher := docker.NewPusher()
	if err := pusher.Push(registryRef); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	fmt.Printf("✓ Successfully pushed %s\n", registryRef)
	return nil
}

// getFilesFromLocalBuild returns list of files in a local build (relative paths)
func getFilesFromLocalBuild(localPath string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read local build directory: %w", err)
	}

	for _, entry := range entries {
		// Skip metadata files
		if entry.Name() == ".aigogo-metadata.json" {
			continue
		}

		if entry.IsDir() {
			// Recursively add files from subdirectories
			subFiles, err := getFilesFromDir(localPath+"/"+entry.Name(), entry.Name())
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// getFilesFromDir recursively gets files from a directory
func getFilesFromDir(fullPath, relPath string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		entryRelPath := relPath + "/" + entry.Name()
		if entry.IsDir() {
			subFiles, err := getFilesFromDir(fullPath+"/"+entry.Name(), entryRelPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			files = append(files, entryRelPath)
		}
	}

	return files, nil
}
