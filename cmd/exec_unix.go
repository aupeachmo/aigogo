//go:build !windows

package cmd

import "syscall"

// replaceProcess replaces the current process with the given binary.
// On Unix, this uses syscall.Exec which ensures signals, exit codes,
// and stdio are properly forwarded with zero overhead.
func replaceProcess(binary string, args []string, env []string) error {
	return syscall.Exec(binary, args, env)
}
