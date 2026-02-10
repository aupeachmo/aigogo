package cmd

import (
	"fmt"
	"runtime"
)

// version is set by main package from build-time ldflags
var version = "0.0.1" // Default fallback

// SetVersion sets the version from main package
func SetVersion(v string) {
	if v != "" {
		version = v
	}
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}

func versionCmd() *Command {
	return &Command{
		Name:        "version",
		Description: "Show aigg version information",
		Run: func(args []string) error {
			fmt.Printf("aigg version %s\n", version)
			fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("  Go version: %s\n", runtime.Version())
			fmt.Println()
			fmt.Println("AI agent manager")
			fmt.Println("https://github.com/aupeachmo/aigogo")
			return nil
		},
	}
}
