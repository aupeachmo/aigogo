package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Flags       *flag.FlagSet
	Run         func(args []string) error
}

// Execute runs the root command
func Execute() error {
	commands := map[string]*Command{
		"init":       initCmd(),
		"add":        addCmd(),
		"install":    installCmd(),
		"rm":         rmCmd(),
		"validate":   validateCmd(),
		"scan":       scanCmd(),
		"build":      buildCmd(),
		"diff":       diffCmd(),
		"push":       pushCmd(),
		"pull":       pullCmd(),
		"login":      loginCmd(),
		"logout":     logoutCmd(),
		"list":       listCmd(),
		"show-deps":  showDepsCmd(),
		"remove":     removeCmd(),
		"remove-all": removeAllCmd(),
		"delete":     deleteCmd(),
		"uninstall":  uninstallCmd(),
		"search":     searchCmd(),
		"version":    versionCmd(),
		"completion": completionCmd(),
	}

	args := os.Args[1:]

	if len(args) == 0 {
		printUsage(commands)
		return nil
	}

	cmdName := args[0]

	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Printf("Unknown command: %s\n\n", cmdName)
		printUsage(commands)
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Parse flags if command has them
	if cmd.Flags != nil {
		// Separate flags and positional args manually
		// This allows flags to appear anywhere (before or after positional args)
		var flagArgs []string
		var posArgs []string

		for i := 1; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "-") {
				// This is a flag
				flagArgs = append(flagArgs, arg)
				// Check if next arg is the flag value (doesn't start with -)
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					flagArgs = append(flagArgs, args[i])
				}
			} else {
				// This is a positional arg
				posArgs = append(posArgs, arg)
			}
		}

		// Parse flags first
		if err := cmd.Flags.Parse(flagArgs); err != nil {
			return err
		}

		// Then add any remaining args from flag parsing
		posArgs = append(posArgs, cmd.Flags.Args()...)
		args = posArgs
	} else {
		args = args[1:]
	}

	return cmd.Run(args)
}

func printUsage(commands map[string]*Command) {
	fmt.Println("aigg - Easily manage and reuse your AI agents between projects")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aigg <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")

	// Define order for better UX
	order := []string{"init", "add", "install", "uninstall", "rm", "validate", "scan", "build", "diff", "push", "pull", "list", "show-deps", "remove", "remove-all", "delete", "login", "logout", "search", "version", "completion"}

	for _, name := range order {
		if cmd, ok := commands[name]; ok {
			fmt.Printf("  %-12s %s\n", name, cmd.Description)
		}
	}

	fmt.Println()
	fmt.Println("Workflow (package consumer):")
	fmt.Println("  aigg add docker.io/org/my-utils:1.0.0  # Add package to aigogo.lock")
	fmt.Println("  aigg install                           # Create import links")
	fmt.Println("  # Python: from aigogo.my_utils import ...")
	fmt.Println("  # JS: import ... from '@aigogo/my-utils'")
	fmt.Println()
	fmt.Println("Workflow (package author):")
	fmt.Println("  aigg init                              # Create aigogo.json")
	fmt.Println("  aigg build utils:1.0.0                 # Build locally")
	fmt.Println("  aigg push docker.io/org/utils:1.0.0    # Push to registry")
	fmt.Println()
	// fmt.Println("For more information, visit: https://github.com/aupeachmo/aigogo")
	// fmt.Println("For more information, visit: https://github.com/aupeachmo/aigogo")
}
