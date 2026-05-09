package stats_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/lesomnus/tegra-exporter/stats"
)

func TestSupervisor(t *testing.T) {
	t.Run("Start and Close", func(t *testing.T) {
		s := stats.NewSupervisor(t.Context(), func(ctx context.Context) (io.ReadCloser, error) {
			r, w := io.Pipe()
			go func() {
				defer w.Close()
				tick := time.Tick(100 * time.Millisecond)
				for {
					select {
					case <-ctx.Done():
						return
					case <-tick:
						w.Write([]byte("05-10-2026 01:19:23\n"))
					}
				}
			}()
			return r, nil
		})

		var stat *stats.Stat
		s.Listen(func(v *stats.Stat) {
			stat = v
			s.Close()
		})
		s.Start()
		s.Wait()
		if stat == nil {
			t.Fatal("stat is nil")
		}
		if stat.Time != time.Date(2026, 5, 10, 1, 19, 23, 0, time.UTC) {
			t.Fatalf("unexpected time: %v", stat.Time)
		}
	})
	t.Run("reopen on error", func(t *testing.T) {
		q := make(chan struct{}, 3)
		s := stats.NewSupervisor(t.Context(), func(ctx context.Context) (io.ReadCloser, error) {
			q <- struct{}{}
			return nil, io.ErrUnexpectedEOF
		})

		s.Fallback = 10 * time.Millisecond
		s.Start()
		<-q
		<-q
		<-q
		s.Close()
		s.Wait()
	})
}
