//go:build windows

package tools

import "os/exec"

// setProcessGroup is a no-op on Windows; process groups are set up
// differently there and aren't needed for this tool's use case.
func setProcessGroup(cmd *exec.Cmd) {}

// killProcessGroup kills the process directly, since Unix-style process
// groups aren't available on Windows.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
