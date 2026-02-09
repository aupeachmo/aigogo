package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/docker"
)

func removeAllCmd() *Command {
	return &Command{
		Name:        "remove-all",
		Description: "Remove all cached snippet packages",
		Run: func(args []string) error {
			flags := flag.NewFlagSet("remove-all", flag.ContinueOnError)
			force := flags.Bool("force", false, "Skip confirmation prompt")

			if err := flags.Parse(args); err != nil {
				return err
			}

			remover := docker.NewRemover()

			// Get count of packages before removal
			lister := docker.NewLister()
			packages, err := lister.List()
			if err != nil {
				return fmt.Errorf("failed to list packages: %w", err)
			}

			if len(packages) == 0 {
				fmt.Println("No cached packages to remove.")
				return nil
			}

			// Show what will be removed
			fmt.Printf("This will remove %d cached package(s):\n\n", len(packages))
			for _, pkg := range packages {
				fmt.Printf("  • %s\n", pkg)
			}
			fmt.Println()

			// Prompt for confirmation unless --force
			if !*force {
				fmt.Print("Are you sure you want to remove ALL cached packages? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				response = strings.TrimSpace(strings.ToLower(response))
				if response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			// Perform removal
			if err := remover.RemoveAll(); err != nil {
				return fmt.Errorf("failed to remove all packages: %w", err)
			}

			fmt.Printf("Successfully removed %d package(s).\n", len(packages))
			return nil
		},
	}
}
