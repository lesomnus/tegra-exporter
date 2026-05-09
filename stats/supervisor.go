package stats

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/lesomnus/otx/log"
	"github.com/lesomnus/signals"
)

type Supervisor struct {
	signals.Event[*Stat]
	Fallback time.Duration

	open func(ctx context.Context) (io.ReadCloser, error)

	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	stopped bool

	done chan struct{}
}

func NewSupervisor(ctx context.Context, open func(ctx context.Context) (io.ReadCloser, error)) *Supervisor {
	c, cancel := context.WithCancel(ctx)
	return &Supervisor{
		Event:    signals.NewEvent[*Stat](),
		Fallback: 3 * time.Second,

		open: open,

		ctx:    c,
		cancel: cancel,

		done: make(chan struct{}),
	}
}

func (s *Supervisor) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	if s.stopped {
		// Closed with no started, do nothing.
		return nil
	}

	s.started = true
	go func() {
		defer close(s.done)
		s.loop()
	}()
	return nil
}

func (s *Supervisor) prepare() (io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return nil, nil
	}

	r, err := s.open(s.ctx)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *Supervisor) loop() {
	l := log.From(s.ctx)
	for {
		r, err := s.prepare()
		if err != nil {
			l.Warn("open failed", slog.String("err", err.Error()))
		} else if r == nil {
			return
		} else {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				stat := &Stat{}
				if err := Parse(line, stat); err != nil {
					l.Warn("parse failed", slog.String("err", err.Error()))
					continue
				}

				s.Event.Emit(stat)
			}
		}

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(s.Fallback):
			// fallback to restart the command if it exits unexpectedly
			// (e.g. due to a transient error)
			// and avoid busy loop if the command exits immediately
		}
	}
}

func (s *Supervisor) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil
	}

	s.stopped = true
	if !s.started {
		return nil
	}

	s.cancel()
	return nil
}

func (s *Supervisor) Wait() {
	<-s.done
}
