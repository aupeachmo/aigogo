//go:build windows

package cmd

import "fmt"

// replaceProcess is not supported on Windows. The exec command relies on
// syscall.Exec (Unix process replacement) for correct signal forwarding and
// exit code propagation, which has no equivalent on Windows.
func replaceProcess(binary string, args []string, env []string) error {
	return fmt.Errorf("aigg exec is not supported on Windows\n" +
		"This command uses Unix process replacement (exec) which is not available on Windows.\n" +
		"Run your agent directly: python <script> or node <script>")
}
