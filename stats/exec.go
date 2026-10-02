package stats

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func Execute(name string, args ...string) func(ctx context.Context) (io.ReadCloser, error) {
	return ExecuteIn("", name, args...)
}

// ExecuteIn runs the command chrooted to `root` unless it is empty.
// With the host's root mounted in a container, this runs the host's `tegrastats`
// against the host's libraries, `/sys`, and `/dev`.
// `name` is resolved inside `root`, so it must be absolute; see [ValidateRoot].
func ExecuteIn(root string, name string, args ...string) func(ctx context.Context) (io.ReadCloser, error) {
	return func(ctx context.Context) (io.ReadCloser, error) {
		ctx, cancel := context.WithCancel(ctx)
		cmd := exec.CommandContext(ctx, name, args...)
		if root != "" {
			if err := chroot(cmd, root); err != nil {
				cancel()
				return nil, err
			}
		}
		r, err := cmd.StdoutPipe()
		if err != nil {
			cancel()
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			cancel()
			return nil, err
		}
		return cmdReadCloser{
			ReadCloser: r,
			cancel:     cancel,
			cmd:        cmd,
		}, nil
	}
}

// ValidateRoot reports whether [ExecuteIn] can run `name` in `root`.
func ValidateRoot(root string, name string) error {
	if !filepath.IsAbs(root) {
		return fmt.Errorf("root path must be absolute: %q", root)
	}
	if !filepath.IsAbs(name) {
		return fmt.Errorf("command must be an absolute path when root path is set: %q", name)
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("root path: %w", err)
	}
	if !info.IsDir() {
		return errors.New("root path is not a directory")
	}
	return nil
}

type cmdReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
	cmd    *exec.Cmd
}

func (c cmdReadCloser) Close() error {
	c.ReadCloser.Close()
	c.cancel()
	return c.cmd.Wait()
}
