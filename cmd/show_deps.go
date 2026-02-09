package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/manifest"
)

func showDepsCmd() *Command {
	flags := flag.NewFlagSet("show-deps", flag.ExitOnError)
	format := flags.String("format", "text", "Output format: text, pyproject, poetry, requirements, npm, yarn")

	return &Command{
		Name:        "show-deps",
		Description: "Show dependencies from aigogo.json in various formats",
		Flags:       flags,
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigogo show-deps <path-to-aigogo.json-or-directory> [--format text|pyproject|poetry|requirements|npm|yarn]\n\nExamples:\n  aigogo show-deps aigogo.json\n  aigogo show-deps vendor/my-snippet\n  aigogo show-deps aigogo.json --format pyproject\n  aigogo show-deps . --format requirements\n  aigogo show-deps . --format npm\n  aigogo show-deps . --format yarn")
			}

			targetPath := args[0]

			// Determine if path is a directory or file
			info, err := os.Stat(targetPath)
			if err != nil {
				return fmt.Errorf("failed to access path: %w", err)
			}

			manifestPath := targetPath
			if info.IsDir() {
				manifestPath = filepath.Join(targetPath, "aigogo.json")
			}

			// Load manifest
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}

			// Output based on format
			switch strings.ToLower(*format) {
			case "text":
				return outputText(m)
			case "pyproject", "pep621":
				return outputPyproject(m)
			case "poetry":
				return outputPoetry(m)
			case "requirements", "pip":
				return outputRequirements(m)
			case "npm", "package-json":
				return outputNpm(m)
			case "yarn":
				return outputYarn(m)
			default:
				return fmt.Errorf("unsupported format: %s\nSupported formats: text, pyproject, poetry, requirements, npm, yarn", *format)
			}
		},
	}
}

func outputText(m *manifest.Manifest) error {
	fmt.Printf("Package: %s\n", m.Name)
	fmt.Printf("Version: %s\n", m.Version)

	if m.Language.Name != "" {
		langStr := m.Language.Name
		if m.Language.Version != "" {
			langStr += " " + m.Language.Version
		}
		fmt.Printf("Language: %s\n", langStr)
	}

	fmt.Println()

	if m.Dependencies == nil || (len(m.Dependencies.Runtime) == 0 && len(m.Dependencies.Dev) == 0) {
		fmt.Println("No dependencies declared")
		return nil
	}

	if len(m.Dependencies.Runtime) > 0 {
		fmt.Printf("Runtime Dependencies (%d):\n", len(m.Dependencies.Runtime))
		for _, dep := range m.Dependencies.Runtime {
			optional := ""
			if dep.Optional {
				optional = " (optional)"
			}
			fmt.Printf("  • %s %s%s\n", dep.Package, dep.Version, optional)
		}
		fmt.Println()
	}

	if len(m.Dependencies.Dev) > 0 {
		fmt.Printf("Development Dependencies (%d):\n", len(m.Dependencies.Dev))
		for _, dep := range m.Dependencies.Dev {
			optional := ""
			if dep.Optional {
				optional = " (optional)"
			}
			fmt.Printf("  • %s %s%s\n", dep.Package, dep.Version, optional)
		}
	}

	return nil
}

func outputPyproject(m *manifest.Manifest) error {
	if strings.ToLower(m.Language.Name) != "python" {
		return fmt.Errorf("pyproject format is only supported for Python packages (current language: %s)", m.Language.Name)
	}

	fmt.Println("# Add these to your pyproject.toml")
	fmt.Println()

	if m.Language.Version != "" {
		fmt.Println("[project]")
		fmt.Printf("requires-python = \"%s\"\n", m.Language.Version)
		fmt.Println()
	}

	if m.Dependencies != nil && len(m.Dependencies.Runtime) > 0 {
		fmt.Println("[project.dependencies]")
		for _, dep := range m.Dependencies.Runtime {
			fmt.Printf("    \"%s%s\",\n", dep.Package, dep.Version)
		}
		fmt.Println()
	}

	if m.Dependencies != nil && len(m.Dependencies.Dev) > 0 {
		fmt.Println("[project.optional-dependencies]")
		fmt.Println("dev = [")
		for _, dep := range m.Dependencies.Dev {
			fmt.Printf("    \"%s%s\",\n", dep.Package, dep.Version)
		}
		fmt.Println("]")
	}

	return nil
}

func outputPoetry(m *manifest.Manifest) error {
	if strings.ToLower(m.Language.Name) != "python" {
		return fmt.Errorf("poetry format is only supported for Python packages (current language: %s)", m.Language.Name)
	}

	fmt.Println("# Add these to your pyproject.toml")
	fmt.Println()

	fmt.Println("[tool.poetry.dependencies]")
	if m.Language.Version != "" {
		fmt.Printf("python = \"%s\"\n", m.Language.Version)
	}

	if m.Dependencies != nil && len(m.Dependencies.Runtime) > 0 {
		for _, dep := range m.Dependencies.Runtime {
			fmt.Printf("%s = \"%s\"\n", dep.Package, dep.Version)
		}
	}
	fmt.Println()

	if m.Dependencies != nil && len(m.Dependencies.Dev) > 0 {
		fmt.Println("[tool.poetry.group.dev.dependencies]")
		for _, dep := range m.Dependencies.Dev {
			fmt.Printf("%s = \"%s\"\n", dep.Package, dep.Version)
		}
	}

	return nil
}

func outputRequirements(m *manifest.Manifest) error {
	if strings.ToLower(m.Language.Name) != "python" {
		return fmt.Errorf("requirements format is only supported for Python packages (current language: %s)", m.Language.Name)
	}

	if m.Dependencies == nil || len(m.Dependencies.Runtime) == 0 {
		fmt.Println("# No runtime dependencies")
		return nil
	}

	fmt.Println("# Runtime dependencies for requirements.txt")
	for _, dep := range m.Dependencies.Runtime {
		fmt.Printf("%s%s\n", dep.Package, dep.Version)
	}

	return nil
}

func isJavaScript(lang string) bool {
	l := strings.ToLower(lang)
	return l == "javascript" || l == "typescript" || l == "js" || l == "ts"
}

func npmVersionRange(ver string) string {
	if ver == "" {
		return "*"
	}
	// Convert Python-style version ranges to npm-compatible semver
	// ">=1.0.0,<2.0.0" → ">=1.0.0 <2.0.0"
	return strings.ReplaceAll(ver, ",", " ")
}

func outputNpm(m *manifest.Manifest) error {
	if !isJavaScript(m.Language.Name) {
		return fmt.Errorf("npm format is only supported for JavaScript/TypeScript packages (current language: %s)", m.Language.Name)
	}

	hasRuntime := m.Dependencies != nil && len(m.Dependencies.Runtime) > 0
	hasDev := m.Dependencies != nil && len(m.Dependencies.Dev) > 0

	fmt.Println("{")

	if hasRuntime {
		fmt.Println("  \"dependencies\": {")
		for i, dep := range m.Dependencies.Runtime {
			comma := ","
			if i == len(m.Dependencies.Runtime)-1 {
				comma = ""
			}
			fmt.Printf("    \"%s\": \"%s\"%s\n", dep.Package, npmVersionRange(dep.Version), comma)
		}
		if hasDev {
			fmt.Println("  },")
		} else {
			fmt.Println("  }")
		}
	}

	if hasDev {
		fmt.Println("  \"devDependencies\": {")
		for i, dep := range m.Dependencies.Dev {
			comma := ","
			if i == len(m.Dependencies.Dev)-1 {
				comma = ""
			}
			fmt.Printf("    \"%s\": \"%s\"%s\n", dep.Package, npmVersionRange(dep.Version), comma)
		}
		fmt.Println("  }")
	}

	fmt.Println("}")

	return nil
}

func outputYarn(m *manifest.Manifest) error {
	if !isJavaScript(m.Language.Name) {
		return fmt.Errorf("yarn format is only supported for JavaScript/TypeScript packages (current language: %s)", m.Language.Name)
	}

	if m.Dependencies == nil || (len(m.Dependencies.Runtime) == 0 && len(m.Dependencies.Dev) == 0) {
		fmt.Println("# No dependencies")
		return nil
	}

	if len(m.Dependencies.Runtime) > 0 {
		fmt.Print("yarn add")
		for _, dep := range m.Dependencies.Runtime {
			fmt.Printf(" \"%s@%s\"", dep.Package, npmVersionRange(dep.Version))
		}
		fmt.Println()
	}

	if len(m.Dependencies.Dev) > 0 {
		fmt.Print("yarn add --dev")
		for _, dep := range m.Dependencies.Dev {
			fmt.Printf(" \"%s@%s\"", dep.Package, npmVersionRange(dep.Version))
		}
		fmt.Println()
	}

	return nil
}
