package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/auth"
	"golang.org/x/term"
)

func loginCmd() *Command {
	flags := flag.NewFlagSet("login", flag.ExitOnError)
	username := flags.String("u", "", "Username")
	passwordStdin := flags.Bool("p", false, "Read password from stdin (prevents password in shell history)")
	dockerhub := flags.Bool("dockerhub", false, "Use Docker Hub (docker.io) as registry")

	return &Command{
		Name:        "login",
		Description: "Login to a container registry",
		Flags:       flags,
		Run: func(args []string) error {
			var registry string
			if *dockerhub {
				registry = "docker.io"
			} else if len(args) < 1 {
				return fmt.Errorf("usage: aigg login <registry> [-u username] [-p] [--dockerhub]")
			} else {
				registry = args[0]
			}

			var user, pass string

			// Get username
			if *username != "" {
				user = *username
			} else {
				fmt.Print("Username: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				user = strings.TrimSpace(input)
			}

			// Get password
			if *passwordStdin {
				// Read password from stdin (for piping or security)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				pass = strings.TrimSpace(input)
			} else {
				// Prompt for password securely (hidden input)
				fmt.Print("Password: ")
				passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				pass = string(passwordBytes)
				fmt.Println() // New line after password input
			}

			// Store credentials
			authManager := auth.NewManager()
			if err := authManager.Login(registry, user, pass); err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			fmt.Printf("Successfully logged in to %s\n", registry)
			return nil
		},
	}
}

func logoutCmd() *Command {
	return &Command{
		Name:        "logout",
		Description: "Logout from a container registry",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigg logout <registry>")
			}

			registry := args[0]

			authManager := auth.NewManager()
			if err := authManager.Logout(registry); err != nil {
				return fmt.Errorf("logout failed: %w", err)
			}

			fmt.Printf("Successfully logged out from %s\n", registry)
			return nil
		},
	}
}
