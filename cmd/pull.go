package cmd

import (
	"fmt"

	"github.com/aupeachmo/aigogo/pkg/docker"
)

func pullCmd() *Command {
	return &Command{
		Name:        "pull",
		Description: "Pull a snippet package from a registry (without extracting)",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigg pull <registry>/<name>:<tag>")
			}

			imageRef := args[0]

			fmt.Printf("Pulling %s...\n", imageRef)

			puller := docker.NewPuller()
			if err := puller.Pull(imageRef); err != nil {
				return fmt.Errorf("failed to pull image: %w", err)
			}

			fmt.Printf("Successfully pulled %s\n", imageRef)
			fmt.Println("Use 'aigg add' + 'aigg install' to set up import links")

			return nil
		},
	}
}
