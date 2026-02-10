package cmd

import (
	"fmt"
)

func searchCmd() *Command {
	return &Command{
		Name:        "search",
		Description: "Search for agents in a registry",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigg search <term>")
			}

			// Note: This is a placeholder. Actual registry search would require
			// implementing the Docker Registry HTTP API V2 search endpoint.
			// Most registries don't support search via API without authentication
			// and additional complexity.

			fmt.Println("Search functionality coming soon!")
			fmt.Println("For now, search directly on your registry's web interface.")
			// fmt.Println("  - Package Registry (e.g., Docker Hub, GHCR, GitLab, etc.)")

			return nil
		},
	}
}
