package cmd

import (
	"fmt"

	"github.com/aupeachmo/aigogo/pkg/depgen"
	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func scanCmd() *Command {
	return &Command{
		Name:        "scan",
		Description: "Scan source files and suggest dependencies",
		Run: func(args []string) error {
			// Load manifest
			m, err := manifest.Load("aigogo.json")
			if err != nil {
				return fmt.Errorf("failed to load aigogo.json: %w\nRun 'aigogo init' first", err)
			}

			fmt.Println("Scanning source files for imports...")
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
				fmt.Println("No files found to scan")
				return nil
			}

			// Scan for imports
			scanner := depgen.NewScanner()
			imports, err := scanner.ScanFiles(files, m.Language.Name)
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			if len(imports) == 0 {
				fmt.Println("No external dependencies detected")
				return nil
			}

			fmt.Printf("Found %d external dependencies:\n", len(imports))
			for _, imp := range imports {
				fmt.Printf("  - %s (in %s)\n", imp.Package, imp.SourceFile)
			}
			fmt.Println()

			// Get currently declared dependencies
			declaredDeps := make(map[string]bool)
			if m.Dependencies != nil {
				for _, dep := range m.Dependencies.Runtime {
					declaredDeps[dep.Package] = true
				}
			}

			// Find missing
			var missing []string
			for _, imp := range imports {
				if !declaredDeps[imp.Package] {
					missing = append(missing, imp.Package)
				}
			}

			if len(missing) > 0 {
				fmt.Println("💡 Add these to your aigogo.json dependencies:")
				fmt.Println()
				for _, pkg := range missing {
					suggestion := suggestDependency(pkg, m.Language.Name)
					fmt.Printf("  %s\n", suggestion)
				}
			} else {
				fmt.Println("✅ All detected dependencies are already declared!")
			}

			return nil
		},
	}
}

func suggestDependency(pkg, lang string) string {
	switch lang {
	case "python":
		return fmt.Sprintf(`{"package": "%s", "version": ">=1.0.0,<2.0.0"}`, pkg)
	case "javascript":
		return fmt.Sprintf(`{"package": "%s", "version": "^1.0.0"}`, pkg)
	case "go":
		return fmt.Sprintf(`{"package": "%s", "version": "v1.0.0"}`, pkg)
	case "rust":
		return fmt.Sprintf(`{"package": "%s", "version": "1.0"}`, pkg)
	}
	return fmt.Sprintf(`{"package": "%s", "version": "1.0.0"}`, pkg)
}
