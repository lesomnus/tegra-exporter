package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/lesomnus/mkot"
	"github.com/lesomnus/otx"
	nooplog "go.opentelemetry.io/otel/log/noop"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	_ "github.com/lesomnus/mkot/otlp"
	"github.com/lesomnus/mkot/pretty"
	_ "github.com/lesomnus/mkot/pretty"
)

type Config struct {
	Stat []string
	Otel OtelConfig
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

	const ServiceResourceId mkot.Id = "resource/tegra-exporter"
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
	// otc.Processors[ServiceResourceId] = &mkot.Resource{
	// 	Attributes: []mkot.Attr{
	// 		{Key: "service.name", Value: attribute.StringValue("tegra-exporter")},
	// 		{Key: "service.version", Value: attribute.StringValue("0.0.0")},
	// 	},
	// }
	if len(otc.Providers) == 0 {
		id := mkot.Id("pretty")
		if _, ok := otc.Exporters[id]; !ok {
			otc.Exporters[id] = pretty.ExporterConfig{}
		}
		otc.Providers["logger"] = &mkot.ProviderConfig{
			Exporters: []mkot.Id{id},
		}
	}

	// for k := range otc.Providers {
	// 	otc.Providers[k].Processors = append(otc.Providers[k].Processors, ServiceResourceId)
	// }

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
