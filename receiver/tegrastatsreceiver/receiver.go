package tegrastatsreceiver

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/lesomnus/otx/log"
	"github.com/lesomnus/tegra-exporter/stats"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

// source runs `tegrastats` (or the fake) and hands each parsed line to a callback.
type source struct {
	command   []string
	root_path string
	logger    *zap.Logger

	supervisor *stats.Supervisor
	stop       func()
}

func (s *source) start(f func(v *stats.Stat)) error {
	if len(s.command) == 1 && s.command[0] == "$fake" {
		s.logger.Warn("use fake stats")
		s.stop = stats.NewFake().Listen(f)
		return nil
	}

	s.logger.Info("run tegrastats",
		zap.Strings("command", s.command),
		zap.String("root_path", s.root_path),
	)

	// The context given to Start must not outlive it, so the supervisor gets its own.
	ctx := log.Into(context.Background(), slog.New(zapslog.NewHandler(s.logger.Core())))
	s.supervisor = stats.NewSupervisor(ctx, stats.ExecuteIn(s.root_path, s.command[0], s.command[1:]...))
	s.stop = s.supervisor.Listen(f)
	return s.supervisor.Start()
}

func (s *source) shutdown() {
	if s.supervisor != nil {
		s.supervisor.Close()
		s.supervisor.Wait()
	}
	if s.stop != nil {
		s.stop()
	}
}

type pushReceiver struct {
	source
	next   consumer.Metrics
	obsrep *receiverhelper.ObsReport
}

func newPushReceiver(c *Config, set receiver.Settings, next consumer.Metrics) (*pushReceiver, error) {
	obsrep, err := receiverhelper.NewObsReport(receiverhelper.ObsReportSettings{
		ReceiverID:             set.ID,
		ReceiverCreateSettings: set,
	})
	if err != nil {
		return nil, err
	}
	return &pushReceiver{
		source: source{command: c.Command, root_path: c.RootPath, logger: set.Logger},
		next:   next,
		obsrep: obsrep,
	}, nil
}

func (r *pushReceiver) Start(_ context.Context, _ component.Host) error {
	return r.source.start(r.consume)
}

func (r *pushReceiver) Shutdown(_ context.Context) error {
	r.source.shutdown()
	return nil
}

func (r *pushReceiver) consume(v *stats.Stat) {
	md := toMetrics(v, pcommon.NewTimestampFromTime(time.Now()))

	ctx := r.obsrep.StartMetricsOp(context.Background())
	err := r.next.ConsumeMetrics(ctx, md)
	r.obsrep.EndMetricsOp(ctx, "tegrastats", md.DataPointCount(), err)
	if err != nil {
		r.logger.Warn("consume metrics", zap.Error(err))
	}
}

type received struct {
	stat *stats.Stat
	at   time.Time
}

type statScraper struct {
	source
	stale_timeout time.Duration

	latest atomic.Pointer[received]
}

func newScraper(c *Config, logger *zap.Logger) *statScraper {
	return &statScraper{
		source:        source{command: c.Command, root_path: c.RootPath, logger: logger},
		stale_timeout: c.StaleTimeout,
	}
}

func (s *statScraper) start(_ context.Context, _ component.Host) error {
	return s.source.start(func(v *stats.Stat) {
		s.latest.Store(&received{stat: v, at: time.Now()})
	})
}

func (s *statScraper) shutdown(_ context.Context) error {
	s.source.shutdown()
	return nil
}

// scrape returns nothing rather than an error when there is no fresh line:
// the supervisor already logs why `tegrastats` is not producing.
func (s *statScraper) scrape(_ context.Context) (pmetric.Metrics, error) {
	v := s.latest.Load()
	if v == nil || time.Since(v.at) > s.stale_timeout {
		return pmetric.NewMetrics(), nil
	}
	return toMetrics(v.stat, pcommon.NewTimestampFromTime(time.Now())), nil
}
