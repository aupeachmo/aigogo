package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func cleanCmd() *Command {
	flags := flag.NewFlagSet("clean", flag.ContinueOnError)
	cleanEnvs := flags.Bool("envs", false, "Remove all exec environments (~/.aigogo/envs/)")
	cleanCache := flags.Bool("cache", false, "Remove build/pull cache (~/.aigogo/cache/)")
	cleanStore := flags.Bool("store", false, "Remove content-addressable store (~/.aigogo/store/)")
	cleanAll := flags.Bool("all", false, "Remove envs, cache, and store")

	return &Command{
		Name:        "clean",
		Description: "Show disk usage or clean cached data",
		Flags:       flags,
		Run: func(args []string) error {
			// If no flags specified, show disk usage summary
			if !*cleanEnvs && !*cleanCache && !*cleanStore && !*cleanAll {
				return showDiskUsage()
			}

			if *cleanAll {
				*cleanEnvs = true
				*cleanCache = true
				*cleanStore = true
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			baseDir := filepath.Join(home, ".aigogo")

			if *cleanEnvs {
				dir := filepath.Join(baseDir, "envs")
				if err := cleanDirectory(dir, "exec environments"); err != nil {
					return err
				}
			}

			if *cleanCache {
				dir := filepath.Join(baseDir, "cache")
				if err := cleanDirectory(dir, "build/pull cache"); err != nil {
					return err
				}
			}

			if *cleanStore {
				dir := filepath.Join(baseDir, "store")
				if err := cleanDirectory(dir, "content-addressable store"); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// showDiskUsage displays the size of each aigogo directory
func showDiskUsage() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(home, ".aigogo")

	type dirInfo struct {
		name string
		path string
		desc string
	}

	dirs := []dirInfo{
		{"Exec environments", filepath.Join(baseDir, "envs"), "aigg clean --envs"},
		{"Build/pull cache", filepath.Join(baseDir, "cache"), "aigg clean --cache"},
		{"Package store", filepath.Join(baseDir, "store"), "aigg clean --store"},
	}

	fmt.Println("aigogo disk usage:")
	fmt.Println()

	var totalSize int64
	for _, d := range dirs {
		size, count := dirStats(d.path)
		totalSize += size
		if size > 0 {
			fmt.Printf("  %-22s %8s  (%d items)   %s\n", d.name+":", formatSize(size), count, d.desc)
		} else {
			fmt.Printf("  %-22s %8s\n", d.name+":", "empty")
		}
	}

	fmt.Println()
	fmt.Printf("  Total:                 %s\n", formatSize(totalSize))
	fmt.Println()
	fmt.Println("Use aigg clean --all to remove everything")

	return nil
}

// cleanDirectory removes a directory and reports what was removed
func cleanDirectory(dir, name string) error {
	size, count := dirStats(dir)
	if size == 0 {
		fmt.Printf("No %s to clean\n", name)
		return nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove %s: %w", name, err)
	}

	fmt.Printf("Removed %s: %s (%d items)\n", name, formatSize(size), count)
	return nil
}

// dirStats returns the total size in bytes and number of top-level items in a directory
func dirStats(path string) (int64, int) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return 0, 0
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, 0
	}

	var totalSize int64
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, 0
	}

	return totalSize, len(entries)
}

// formatSize is defined in list.go
