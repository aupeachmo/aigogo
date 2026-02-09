package cmd

import (
	"fmt"

	"github.com/aupeachmo/aigogo/pkg/depgen"
	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func validateCmd() *Command {
	return &Command{
		Name:        "validate",
		Description: "Validate dependencies against actual imports in source files",
		Run: func(args []string) error {
			// Load manifest
			m, err := manifest.Load("aigogo.json")
			if err != nil {
				return fmt.Errorf("failed to load aigogo.json: %w\nRun 'aigogo init' first", err)
			}

			fmt.Println("Validating manifest...")
			fmt.Println()

			// Discover files
			discovery, err := manifest.NewFileDiscovery(".", m.Files.Exclude)
			if err != nil {
				return fmt.Errorf("failed to initialize file discovery: %w", err)
			}
			files, err := discovery.Discover(m.Files, m.Language)
			if err != nil {
				return fmt.Errorf("failed to discover files: %w", err)
			}

			if len(files) == 0 {
				return fmt.Errorf("no files found to validate")
			}

			fmt.Printf("📁 Found %d file(s) to scan\n", len(files))
			fmt.Println()

			// Validate dependencies
			validator := depgen.NewValidator()
			result, err := validator.Validate(m, files)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			// Print detected imports
			if len(result.Imports) > 0 {
				fmt.Println("📦 Detected external imports:")
				for _, imp := range result.Imports {
					fmt.Printf("  - %s (in %s:%d)\n", imp.Package, imp.SourceFile, imp.LineNumber)
				}
				fmt.Println()
			}

			// Print errors
			if len(result.Errors) > 0 {
				fmt.Println("❌ Errors:")
				for _, err := range result.Errors {
					fmt.Printf("  - %s\n", err)
				}
				fmt.Println()
			}

			// Print warnings
			if len(result.Warnings) > 0 {
				fmt.Println("⚠️  Warnings:")
				for _, warn := range result.Warnings {
					fmt.Printf("  - %s\n", warn)
				}
				fmt.Println()
			}

			// Print suggestions
			if len(result.Suggestions) > 0 {
				fmt.Println("💡 Suggestions:")
				for _, sug := range result.Suggestions {
					fmt.Printf("  %s\n", sug)
				}
				fmt.Println()
			}

			// Final status
			if result.Valid {
				fmt.Println("✅ Validation passed!")
				return nil
			}

			fmt.Println("❌ Validation failed")
			fmt.Println("Fix the issues above before pushing")
			return fmt.Errorf("validation failed")
		},
	}
}
