//go:build unix

package stats

import (
	"os/exec"
	"syscall"
)

func chroot(cmd *exec.Cmd, root string) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: root}
	// Otherwise the working directory stays outside of the new root.
	cmd.Dir = "/"
	return nil
}
