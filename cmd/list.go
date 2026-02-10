package cmd

import (
	"fmt"
	"time"

	"github.com/aupeachmo/aigogo/pkg/docker"
)

func listCmd() *Command {
	return &Command{
		Name:        "list",
		Description: "List cached snippet packages",
		Run: func(args []string) error {
			lister := docker.NewLister()
			images, err := lister.ListDetailed()
			if err != nil {
				return fmt.Errorf("failed to list images: %w", err)
			}

			if len(images) == 0 {
				fmt.Println("No cached snippet packages found")
				fmt.Println("\nTip: Build a local package with: aigg build <name>:<tag>")
				fmt.Println("     Or add from registry with:  aigg add <registry>/<name>:<tag>")
				return nil
			}

			fmt.Printf("Cached snippet packages (%d):\n\n", len(images))

			for _, img := range images {
				// Format the image line
				typeIndicator := "📦"
				if img.Type == "local-build" {
					typeIndicator = "🔨"
				}

				fmt.Printf("%s %s\n", typeIndicator, img.Name)

				// Show type
				typeLabel := "local build"
				if img.Type == "registry-pull" {
					typeLabel = "pulled from registry"
				}
				fmt.Printf("   Type: %s\n", typeLabel)

				// Show time
				timeAgo := formatTimeAgo(img.BuildTime)
				fmt.Printf("   Time: %s\n", timeAgo)

				// Show size
				sizeStr := formatSize(img.Size)
				fmt.Printf("   Size: %s\n", sizeStr)

				// Show language and version if manifest is available
				if img.Manifest != nil {
					if img.Manifest.Language.Name != "" {
						langStr := img.Manifest.Language.Name
						if img.Manifest.Language.Version != "" {
							langStr += " " + img.Manifest.Language.Version
						}
						fmt.Printf("   Language: %s\n", langStr)
					}

					// Show dependency count
					if img.Manifest.Dependencies != nil {
						runtimeCount := len(img.Manifest.Dependencies.Runtime)
						devCount := len(img.Manifest.Dependencies.Dev)

						if runtimeCount > 0 || devCount > 0 {
							depStr := fmt.Sprintf("%d runtime", runtimeCount)
							if devCount > 0 {
								depStr += fmt.Sprintf(", %d dev", devCount)
							}
							fmt.Printf("   Dependencies: %s\n", depStr)
						}
					}
				}

				fmt.Println()
			}

			return nil
		},
	}
}

// formatTimeAgo formats a time as "X ago" string
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// formatSize formats bytes as human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
