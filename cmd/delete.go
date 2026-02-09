package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aupeach/aigogo/pkg/docker"
)

func deleteCmd() *Command {
	// Create flag set for the delete command
	flags := flag.NewFlagSet("delete", flag.ContinueOnError)
	all := flags.Bool("all", false, "Delete all tags in the repository")

	return &Command{
		Name:        "delete",
		Description: "Delete a snippet package from remote registry",
		Flags:       flags,
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigogo delete <registry>/<name>:<tag> [--all]")
			}

			imageRef := args[0]
			deleter := docker.NewDeleter()

			if *all {
				// Delete all tags in the repository
				// Parse the image ref to extract registry and repository (ignore tag)
				parts := strings.SplitN(imageRef, "/", 2)
				if len(parts) < 2 {
					return fmt.Errorf("invalid image reference format (expected: registry/repository or registry/repository:tag)")
				}

				registry := parts[0]
				repository := parts[1]

				// Remove tag if present (we don't need it for --all)
				if idx := strings.LastIndex(repository, ":"); idx != -1 {
					repository = repository[:idx]
				}

				fmt.Printf("⚠️  WARNING: This will permanently delete ALL tags in %s/%s\n\n", registry, repository)

				// List tags first (to show what will be deleted)
				// This is done inside DeleteAll, but we want to confirm first
				fmt.Print("Type 'DELETE ALL' to confirm: ")

				// Use bufio.Reader to read the entire line (including spaces)
				reader := bufio.NewReader(os.Stdin)
				confirmation, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				confirmation = strings.TrimSpace(confirmation)

				if confirmation != "DELETE ALL" {
					fmt.Println("Delete cancelled")
					return nil
				}

				fmt.Println()
				if err := deleter.DeleteAll(registry, repository); err != nil {
					return fmt.Errorf("failed to delete all tags: %w", err)
				}

			} else {
				// Delete a specific tag
				fmt.Printf("⚠️  WARNING: This will permanently delete %s from the registry\n", imageRef)
				fmt.Print("Are you sure? Type 'yes' to confirm: ")

				var confirmation string
				_, _ = fmt.Scanln(&confirmation)

				if strings.ToLower(confirmation) != "yes" {
					fmt.Println("Delete cancelled")
					return nil
				}

				fmt.Printf("Deleting %s from registry...\n", imageRef)

				if err := deleter.Delete(imageRef); err != nil {
					return fmt.Errorf("failed to delete: %w", err)
				}

				fmt.Printf("✓ Successfully deleted %s from registry\n", imageRef)
				fmt.Println()
				fmt.Println("Note: The local cache is not affected. To remove from cache, run:")
				fmt.Printf("  aigogo remove %s\n", imageRef)
			}

			return nil
		},
	}
}
