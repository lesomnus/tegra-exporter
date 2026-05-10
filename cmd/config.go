package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/lesomnus/mkot"
	"github.com/lesomnus/otx"
	"github.com/lesomnus/otx/log"
	"github.com/lesomnus/tegra-exporter/cmd/version"
	"github.com/lesomnus/xli"
	"go.opentelemetry.io/otel/attribute"
	nooplog "go.opentelemetry.io/otel/log/noop"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	_ "github.com/lesomnus/mkot/otlp"
	"github.com/lesomnus/mkot/pretty"
	_ "github.com/lesomnus/mkot/pretty"
	"github.com/lesomnus/mkot/prometheus"
	_ "github.com/lesomnus/mkot/prometheus"
)

func NewCmdConfig() *xli.Command {
	return &xli.Command{
		Name: "config",
		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c, rp, err := readConfig()
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			ctx, _, err = c.Otel.Build(ctx)
			if err != nil {
				return fmt.Errorf("build otel: %w", err)
			}

			l := log.From(ctx)
			l.Info("config loaded", slog.String("path", rp))

			return yaml.NewEncoder(cmd).Encode(c)
		}),
	}
}

type Config struct {
	Stat []string
	Otel OtelConfig
}

func readConfig() (*Config, string, error) {
	path_to_lookup := []string{
		"tegra-exporter.yaml",
		"tegra-exporter.yml",
	}

	var (
		r  io.Reader
		rp string
	)
	for _, rp = range path_to_lookup {
		f, err := os.Open(rp)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, "", fmt.Errorf("open config file: %w", err)
		}
		r = f
		break
	}

	var c Config
	if r != nil {
		if err := yaml.NewDecoder(r).Decode(&c); err != nil {
			return nil, "", fmt.Errorf("decode config: %w", err)
		}
	}
	if err := c.Evaluate(); err != nil {
		return nil, "", fmt.Errorf("evaluate config: %w", err)
	}
	return &c, rp, nil
}

func (c *Config) Evaluate() error {
	if len(c.Stat) == 0 {
		c.Stat = []string{"tegrastats"}
	}
	return nil
}

type OtelConfig struct {
	mkot.Config `yaml:",inline"`
}

func (c *OtelConfig) Build(ctx context.Context) (context.Context, *otx.Otx, error) {
	otc := mkot.NewConfig()
	if c != nil {
		otc = &c.Config
	}

	if otc.Processors == nil {
		otc.Processors = map[mkot.Id]mkot.ProcessorConfig{}
	}
	if otc.Exporters == nil {
		otc.Exporters = map[mkot.Id]mkot.ExporterConfig{}
	}
	if otc.Processors == nil {
		otc.Processors = map[mkot.Id]mkot.ProcessorConfig{}
	}
	if otc.Providers == nil {
		otc.Providers = map[mkot.Id]*mkot.ProviderConfig{}
	}

	const ServiceResourceId mkot.Id = "resource/tegra-exporter"
	if _, ok := otc.Processors[ServiceResourceId]; !ok {
		otc.Processors[ServiceResourceId] = &mkot.Resource{
			Attributes: []mkot.Attr{
				{Key: "service.name", Value: attribute.StringValue("tegra-exporter")},
				{Key: "service.version", Value: attribute.StringValue(version.Get().Version)},
			},
		}
	}
	if _, ok := otc.Exporters["pretty"]; !ok {
		otc.Exporters["pretty"] = pretty.ExporterConfig{}
	}
	if _, ok := otc.Exporters["prometheus/local"]; !ok {
		otc.Exporters["prometheus/local"] = &prometheus.ExporterConfig{
			Namespace: "tegra_exporter",
			Endpoint:  ":8888",
		}
	}
	if _, ok := otc.Providers["logger"]; !ok {
		otc.Providers["logger"] = &mkot.ProviderConfig{
			Exporters: []mkot.Id{"pretty"},
		}
	}
	if _, ok := otc.Providers["meter"]; !ok {
		otc.Providers["meter"] = &mkot.ProviderConfig{
			Exporters:  []mkot.Id{"prometheus/local"},
			Processors: []mkot.Id{ServiceResourceId},
		}
	}

	resolver := mkot.Make(ctx, otc)

	tracker_provider, err := resolver.Tracer(ctx, "")
	if err != nil {
		if !errors.Is(err, mkot.ErrNotExist) {
			return nil, nil, fmt.Errorf("resolve tracer provider: %w", err)
		}
		tracker_provider = nooptrace.NewTracerProvider()
	}

	meter_provider, err := resolver.Meter(ctx, "")
	if err != nil {
		if !errors.Is(err, mkot.ErrNotExist) {
			return nil, nil, fmt.Errorf("resolve meter provider: %w", err)
		}
		meter_provider = noopmetric.NewMeterProvider()
	}

	logger_provider, err := resolver.Logger(ctx, "")
	if err != nil {
		if !errors.Is(err, mkot.ErrNotExist) {
			return nil, nil, fmt.Errorf("resolve logger provider: %w", err)
		}
		logger_provider = nooplog.NewLoggerProvider()
	}
	v := otx.New(
		otx.WithController(resolver),
		otx.WithTracerProvider(tracker_provider),
		otx.WithMeterProvider(meter_provider),
		otx.WithLoggerProvider(logger_provider),
	)
	return otx.Into(ctx, v), v, nil
}
