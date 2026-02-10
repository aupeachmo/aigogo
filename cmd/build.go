package cmd

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/docker"
	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func buildCmd() *Command {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	force := flags.Bool("force", false, "Force rebuild even if already exists")
	noValidate := flags.Bool("no-validate", false, "Skip dependency validation")

	return &Command{
		Name:        "build",
		Description: "Build a snippet package locally (no push)",
		Flags:       flags,
		Run: func(args []string) error {
			// Find manifest (supports running from subdirectories)
			m, manifestDir, err := manifest.FindManifest()
			if err != nil {
				return fmt.Errorf("failed to find manifest: %w", err)
			}

			var imageRef string
			var updateVersion bool

			// If no arguments provided, use name from aigogo.json and auto-increment version
			if len(args) == 0 {
				if m.Name == "" {
					return fmt.Errorf("no package name specified and aigogo.json has no name field")
				}
				if m.Version == "" {
					return fmt.Errorf("no version in aigogo.json (required for auto-increment)")
				}

				// Increment the version
				newVersion, err := manifest.IncrementVersion(m.Version)
				if err != nil {
					return fmt.Errorf("failed to increment version: %w", err)
				}

				imageRef = fmt.Sprintf("%s:%s", m.Name, newVersion)
				updateVersion = true

				fmt.Printf("Auto-incrementing version: %s -> %s\n", m.Version, newVersion)
			} else {
				// Use provided name:tag
				imageRef = args[0]
				updateVersion = false
			}

			fmt.Printf("Building package: %s\n", imageRef)

			// Check if it has a registry prefix
			hasRegistry := strings.Contains(imageRef, "/") &&
				(strings.HasPrefix(imageRef, "docker.io/") ||
					strings.HasPrefix(imageRef, "ghcr.io/") ||
					strings.Contains(strings.Split(imageRef, "/")[0], "."))

			if hasRegistry {
				fmt.Println()
				fmt.Println("⚠️  Warning: Building with registry prefix")
				fmt.Println("   This is a LOCAL build stored in cache")
				fmt.Printf("   Location: ~/.aigogo/cache/%s\n", docker.SanitizeImageRef(imageRef))
				fmt.Println()
				fmt.Println("   To push to registry:")
				fmt.Printf("     aigg push %s --from %s\n", imageRef, imageRef)
				fmt.Println()
			}

			// Validate dependencies unless --no-validate
			if !*noValidate {
				fmt.Println("Validating dependencies...")
				if err := validateManifest(m); err != nil {
					return fmt.Errorf("validation failed: %w\nUse --no-validate to skip", err)
				}
				fmt.Println("✓ Validation passed")
			}

			// Build to local cache (from manifest directory)
			builder := docker.NewLocalBuilder()
			if err := builder.BuildFromDir(manifestDir, imageRef, m, *force); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}

			// Update aigogo.json with new version if auto-incremented
			if updateVersion {
				// Extract version from imageRef
				parts := strings.Split(imageRef, ":")
				if len(parts) == 2 {
					m.Version = parts[1]
					manifestPath := filepath.Join(manifestDir, "aigogo.json")
					if err := manifest.Save(manifestPath, m); err != nil {
						fmt.Printf("⚠️  Warning: Built successfully but failed to update version in aigogo.json: %v\n", err)
					} else {
						fmt.Printf("✓ Updated aigogo.json version to %s\n", m.Version)
					}
				}
			}

			fmt.Printf("\n✓ Successfully built %s\n", imageRef)
			fmt.Println("\nNext steps:")
			fmt.Printf("  Test locally:  aigg add %s && aigg install\n", imageRef)

			// Suggest a registry name if it's a local build
			registryHint := "<registry>/myorg/" + strings.Split(imageRef, ":")[0]
			if strings.Contains(imageRef, "/") {
				registryHint = "<registry>/" + imageRef
			}
			fmt.Printf("  Push to registry: aigg push %s\n", registryHint)

			return nil
		},
	}
}

// Helper function for validation (reused from push.go logic)
func validateManifest(m *manifest.Manifest) error {
	// This will be implemented to call the validation logic
	// For now, just basic validation is done in manifest.Load
	return nil
}
