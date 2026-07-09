//go:build !windows

package tools

import (
	"os/exec"
	"syscall"
)

// setProcessGroup puts cmd in its own process group so killProcessGroup can
// terminate it along with any children (e.g. a grandchild like `sleep`
// spawned by `sh -c`) rather than just the shell itself.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessGroup kills the process group started by setProcessGroup.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
