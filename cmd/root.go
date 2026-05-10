package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lesomnus/otx"
	"github.com/lesomnus/otx/log"
	"github.com/lesomnus/tegra-exporter/stats"
	"github.com/lesomnus/xli"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func NewCmdRoot() *xli.Command {
	return &xli.Command{
		Name: "tegra-exporter",
		Commands: xli.Commands{
			NewCmdVersion(),
			NewCmdConfig(),
		},
		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c, rp, err := readConfig()
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			ctx, otx, err := c.Otel.Build(ctx)
			if err != nil {
				return fmt.Errorf("build otel: %w", err)
			}
			if err := otx.Start(ctx); err != nil {
				return fmt.Errorf("start otel: %w", err)
			}
			defer otx.Shutdown(ctx)

			collector, err := newCollector(ctx)
			if err != nil {
				return fmt.Errorf("create collector: %w", err)
			}

			l := log.From(ctx)
			l.Info("config loaded", slog.String("path", rp))
			l.Info("executing stats", slog.String("args", strings.Join(c.Stat, " ")))

			args := []string{}
			if len(c.Stat) > 1 {
				args = c.Stat[1:]
			}
			execute := stats.Execute(c.Stat[0], args...)
			s := stats.NewSupervisor(ctx, execute)
			if err := s.Start(); err != nil {
				return fmt.Errorf("start supervisor: %w", err)
			}

			stop := s.Listen(collector)
			defer stop()

			s.Wait()
			return next(ctx)
		}),
	}
}

func newCollector(ctx context.Context) (func(v *stats.Stat), error) {
	meter := otx.Meter(ctx)

	errs := []error{}
	ram_inuse := newInt64Gauge(&errs, meter, "tegra.ram.in_use", "RAM in use", "MB")
	ram_total := newInt64Gauge(&errs, meter, "tegra.ram.total", "Total RAM", "MB")
	ram_lfb_count := newInt64Gauge(&errs, meter, "tegra.ram.lfb_count", "Largest Free Block count", "count")
	ram_lfb_size := newInt64Gauge(&errs, meter, "tegra.ram.lfb_size", "Largest Free Block size", "MB")
	iram_inuse := newInt64Gauge(&errs, meter, "tegra.iram.in_use", "IRAM in use", "MB")
	iram_total := newInt64Gauge(&errs, meter, "tegra.iram.total", "Total IRAM", "MB")
	swap_inuse := newInt64Gauge(&errs, meter, "tegra.swap.in_use", "Swap in use", "MB")
	swap_total := newInt64Gauge(&errs, meter, "tegra.swap.total", "Total swap", "MB")
	swap_cached := newInt64Gauge(&errs, meter, "tegra.swap.cached", "Cached swap", "MB")
	cpu_utilization := newInt64Gauge(&errs, meter, "tegra.cpu.utilization", "CPU utilization", "%")
	cpu_frequency := newInt64Gauge(&errs, meter, "tegra.cpu.frequency", "CPU frequency", "MHz")
	emc_utilization := newInt64Gauge(&errs, meter, "tegra.emc.utilization", "EMC utilization", "%")
	emc_frequency := newInt64Gauge(&errs, meter, "tegra.emc.frequency", "EMC frequency", "MHz")
	gr3d_utilization := newInt64Gauge(&errs, meter, "tegra.gr3d.utilization", "GR3D utilization", "%")
	gr3d_frequency := newInt64Gauge(&errs, meter, "tegra.gr3d.frequency", "GR3D frequency", "MHz")
	vic_utilization := newInt64Gauge(&errs, meter, "tegra.vic.utilization", "VIC utilization", "%")
	vic_frequency := newInt64Gauge(&errs, meter, "tegra.vic.frequency", "VIC frequency", "MHz")
	ape_frequency := newInt64Gauge(&errs, meter, "tegra.ape.frequency", "APE frequency", "MHz")
	nvenc_utilization := newInt64Gauge(&errs, meter, "tegra.nvenc.utilization", "NVENC utilization", "%")
	nvenc_frequency := newInt64Gauge(&errs, meter, "tegra.nvenc.frequency", "NVENC frequency", "MHz")
	nvdec_utilization := newInt64Gauge(&errs, meter, "tegra.nvdec.utilization", "NVDEC utilization", "%")
	nvdec_frequency := newInt64Gauge(&errs, meter, "tegra.nvdec.frequency", "NVDEC frequency", "MHz")
	nvdla_utilization := newInt64Gauge(&errs, meter, "tegra.nvdla.utilization", "NVDLA utilization", "%")
	nvdla_frequency := newInt64Gauge(&errs, meter, "tegra.nvdla.frequency", "NVDLA frequency", "MHz")
	nvjpg_utilization := newInt64Gauge(&errs, meter, "tegra.nvjpg.utilization", "NVJPG utilization", "%")
	nvjpg_frequency := newInt64Gauge(&errs, meter, "tegra.nvjpg.frequency", "NVJPG frequency", "MHz")
	// pva_utilization := newInt64Gauge(&errs, meter, "tegra.pva.utilization", "PVA utilization", "%")
	// pva_frequency := newInt64Gauge(&errs, meter, "tegra.pva.frequency", "PVA frequency", "MHz")
	ofa_utilization := newInt64Gauge(&errs, meter, "tegra.ofa.utilization", "OFA utilization", "%")
	ofa_frequency := newInt64Gauge(&errs, meter, "tegra.ofa.frequency", "OFA frequency", "MHz")
	temperature := newInt64Gauge(&errs, meter, "tegra.temperature", "Temperature", "C")
	power_current := newInt64Gauge(&errs, meter, "tegra.power.current", "Power consumption", "mW")
	power_average := newInt64Gauge(&errs, meter, "tegra.power.average", "Average power consumption", "mW")

	if len(errs) > 0 {
		return nil, fmt.Errorf("create instruments: %w", errors.Join(errs...))
	}

	return func(v *stats.Stat) {
		if !IsZero(v.Ram) {
			ram_inuse.Record(ctx, int64(v.Ram.InUse))
			ram_total.Record(ctx, int64(v.Ram.Total))
			ram_lfb_count.Record(ctx, int64(v.Ram.LfbCount))
			ram_lfb_size.Record(ctx, int64(v.Ram.LfbSize))
		}
		if !IsZero(v.IRam) {
			iram_inuse.Record(ctx, int64(v.IRam.InUse))
			iram_total.Record(ctx, int64(v.IRam.Total))
		}
		if !IsZero(v.Swap) {
			swap_inuse.Record(ctx, int64(v.Swap.InUse))
			swap_total.Record(ctx, int64(v.Swap.Total))
			swap_cached.Record(ctx, int64(v.Swap.Cached))
		}
		for i, w := range v.Cpus {
			attr := metric.WithAttributes(attribute.String("index", fmt.Sprintf("%d", i)))
			cpu_utilization.Record(ctx, int64(w.Percent), attr)
			cpu_frequency.Record(ctx, int64(w.Freq), attr)
		}
		if !IsZero(v.Emc) {
			emc_utilization.Record(ctx, int64(v.Emc.Percent))
			emc_frequency.Record(ctx, int64(v.Emc.Freq))
		}
		if v.Gr3d.Freq != nil {
			gr3d_utilization.Record(ctx, int64(v.Gr3d.Percent))
			for i, w := range v.Gr3d.Freq {
				attr := metric.WithAttributes(attribute.String("index", fmt.Sprintf("%d", i)))
				gr3d_frequency.Record(ctx, int64(w), attr)
			}
		}
		if !IsZero(v.Vic) {
			vic_utilization.Record(ctx, int64(v.Vic.Percent))
			vic_frequency.Record(ctx, int64(v.Vic.Freq))
		}
		if !IsZero(v.Ape) {
			ape_frequency.Record(ctx, int64(v.Ape.Freq))
		}
		for i, w := range v.NvEnc {
			attr := metric.WithAttributes(attribute.String("index", fmt.Sprintf("%d", i)))
			nvenc_utilization.Record(ctx, int64(w.Percent), attr)
			nvenc_frequency.Record(ctx, int64(w.Freq), attr)
		}
		for i, w := range v.NvDec {
			attr := metric.WithAttributes(attribute.String("index", fmt.Sprintf("%d", i)))
			nvdec_utilization.Record(ctx, int64(w.Percent), attr)
			nvdec_frequency.Record(ctx, int64(w.Freq), attr)
		}
		for i, w := range v.NvDla {
			attr := metric.WithAttributes(attribute.String("index", fmt.Sprintf("%d", i)))
			nvdla_utilization.Record(ctx, int64(w.Percent), attr)
			nvdla_frequency.Record(ctx, int64(w.Freq), attr)
		}
		for i, w := range v.NvJpg {
			attr := metric.WithAttributes(attribute.String("index", fmt.Sprintf("%d", i)))
			nvjpg_utilization.Record(ctx, int64(w.Percent), attr)
			nvjpg_frequency.Record(ctx, int64(w.Freq), attr)
		}
		// for i, pva := range v.Pva {
		// 	attr := metric.WithAttributes(attribute.String("index", fmt.Sprintf("%d", i)))
		// 	pva_utilization.Record(ctx, int64(pva.Percent), attr)
		// 	pva_frequency.Record(ctx, int64(pva.Freq), attr)
		// }
		if !IsZero(v.Ofa) {
			ofa_utilization.Record(ctx, int64(v.Ofa.Percent))
			ofa_frequency.Record(ctx, int64(v.Ofa.Freq))
		}
		for _, w := range v.Temp {
			attr := metric.WithAttributes(attribute.String("sensor", w.Name))
			temperature.Record(ctx, int64(w.Value), attr)
		}
		for _, w := range v.Power {
			attr := metric.WithAttributes(attribute.String("sensor", w.Name))
			power_current.Record(ctx, int64(w.Current), attr)
			power_average.Record(ctx, int64(w.Average), attr)
		}
	}, nil
}

func newInt64Gauge(errs *[]error, meter metric.Meter, name, desc, unit string) metric.Int64Gauge {
	v, err := meter.Int64Gauge(name, metric.WithDescription(desc), metric.WithUnit(unit))
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s: %w", name, err))
	}
	return v
}

func IsZero[T comparable](v T) bool {
	var zero T
	return v == zero
}
