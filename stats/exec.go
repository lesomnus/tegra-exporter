package stats

import (
	"context"
	"io"
	"os/exec"
)

func Execute(name string, args ...string) func(ctx context.Context) (io.ReadCloser, error) {
	return func(ctx context.Context) (io.ReadCloser, error) {
		ctx, cancel := context.WithCancel(ctx)
		cmd := exec.CommandContext(ctx, name, args...)
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
