package cmd

import (
	"fmt"

	"github.com/aupeachmo/aigogo/pkg/docker"
)

func removeCmd() *Command {
	return &Command{
		Name:        "remove",
		Description: "Remove a cached snippet package",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigg remove <name>:<tag>")
			}

			imageRef := args[0]

			remover := docker.NewRemover()
			if err := remover.Remove(imageRef); err != nil {
				return fmt.Errorf("failed to remove image: %w", err)
			}

			fmt.Printf("Successfully removed %s\n", imageRef)
			return nil
		},
	}
}
