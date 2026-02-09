package main

import (
	"fmt"
	"os"

	"github.com/aupeachmo/aigogo/cmd"
)

// Version is set via ldflags at build time
// Example: go build -ldflags="-X main.Version=v2.0.1"
var Version = "0.0.1"

func main() {
	// Make version available to commands
	cmd.SetVersion(Version)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
