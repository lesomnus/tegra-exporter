//go:build !unix

package stats

import (
	"errors"
	"os/exec"
)

func chroot(cmd *exec.Cmd, root string) error {
	return errors.New("chroot is not supported on this platform")
}
