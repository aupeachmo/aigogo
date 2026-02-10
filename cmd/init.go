package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func initCmd() *Command {
	return &Command{
		Name:        "init",
		Description: "Initialize a new agent in the current directory",
		Run: func(args []string) error {
			manifestPath := "aigogo.json"

			// Create default manifest
			m := &manifest.Manifest{
				Schema:      "https://github.com/aupeachmo/aigogo/blob/master/aigogo.schema.json",
				Name:        getCurrentDirName(),
				Version:     "0.1.0",
				Description: "An AI agent",
				Author:      "",
				Language: manifest.Language{
					Name:    "python",
					Version: ">=3.8,<4.0",
				},
				Dependencies: &manifest.Dependencies{
					Runtime: []manifest.Dependency{},
					Dev:     []manifest.Dependency{},
				},
				Files: manifest.FileSpec{
					Include: []string{},
					Exclude: []string{},
				},
				Metadata: manifest.Metadata{
					License: "MPL-2.0",
					Tags:    []string{},
				},
			}

			// Save manifest
			if err := manifest.Save(manifestPath, m); err != nil {
				return err
			}

			fmt.Println("✓ Initialized aigogo package")
			fmt.Printf("  Created %s\n\n", manifestPath)
			fmt.Println("Next steps:")
			fmt.Println("  1. Edit aigogo.json to configure language and metadata")
			fmt.Println("  2. Add files: aigg add file <path>...")
			fmt.Println("  3. Add dependencies: aigg add dep <package> <version>")
			fmt.Println("  4. Run 'aigg validate' to check your configuration")
			fmt.Println("  5. Build and share: aigg build <name>:<tag>")

			return nil
		},
	}
}

func getCurrentDirName() string {
	dir, err := filepath.Abs(".")
	if err != nil {
		return "snippet-package"
	}
	return filepath.Base(dir)
}
