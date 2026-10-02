package tegrastatsreceiver

import (
	"strconv"

	"github.com/lesomnus/tegra-exporter/stats"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Names, units, and conditions follow `newCollector` in `cmd/root.go`
// so both ways of running produce the same metrics.

type gauge struct {
	name string
	desc string
	unit string
}

var (
	ram_inuse         = gauge{"tegra.ram.in_use", "RAM in use", "MB"}
	ram_total         = gauge{"tegra.ram.total", "Total RAM", "MB"}
	ram_lfb_count     = gauge{"tegra.ram.lfb_count", "Largest Free Block count", "count"}
	ram_lfb_size      = gauge{"tegra.ram.lfb_size", "Largest Free Block size", "MB"}
	iram_inuse        = gauge{"tegra.iram.in_use", "IRAM in use", "MB"}
	iram_total        = gauge{"tegra.iram.total", "Total IRAM", "MB"}
	swap_inuse        = gauge{"tegra.swap.in_use", "Swap in use", "MB"}
	swap_total        = gauge{"tegra.swap.total", "Total swap", "MB"}
	swap_cached       = gauge{"tegra.swap.cached", "Cached swap", "MB"}
	cpu_utilization   = gauge{"tegra.cpu.utilization", "CPU utilization", "%"}
	cpu_frequency     = gauge{"tegra.cpu.frequency", "CPU frequency", "MHz"}
	emc_utilization   = gauge{"tegra.emc.utilization", "EMC utilization", "%"}
	emc_frequency     = gauge{"tegra.emc.frequency", "EMC frequency", "MHz"}
	gr3d_utilization  = gauge{"tegra.gr3d.utilization", "GR3D utilization", "%"}
	gr3d_frequency    = gauge{"tegra.gr3d.frequency", "GR3D frequency", "MHz"}
	vic_utilization   = gauge{"tegra.vic.utilization", "VIC utilization", "%"}
	vic_frequency     = gauge{"tegra.vic.frequency", "VIC frequency", "MHz"}
	ape_frequency     = gauge{"tegra.ape.frequency", "APE frequency", "MHz"}
	nvenc_utilization = gauge{"tegra.nvenc.utilization", "NVENC utilization", "%"}
	nvenc_frequency   = gauge{"tegra.nvenc.frequency", "NVENC frequency", "MHz"}
	nvdec_utilization = gauge{"tegra.nvdec.utilization", "NVDEC utilization", "%"}
	nvdec_frequency   = gauge{"tegra.nvdec.frequency", "NVDEC frequency", "MHz"}
	nvdla_utilization = gauge{"tegra.nvdla.utilization", "NVDLA utilization", "%"}
	nvdla_frequency   = gauge{"tegra.nvdla.frequency", "NVDLA frequency", "MHz"}
	nvjpg_utilization = gauge{"tegra.nvjpg.utilization", "NVJPG utilization", "%"}
	nvjpg_frequency   = gauge{"tegra.nvjpg.frequency", "NVJPG frequency", "MHz"}
	ofa_utilization   = gauge{"tegra.ofa.utilization", "OFA utilization", "%"}
	ofa_frequency     = gauge{"tegra.ofa.frequency", "OFA frequency", "MHz"}
	temperature       = gauge{"tegra.temperature", "Temperature", "C"}
	power_current     = gauge{"tegra.power.current", "Power consumption", "mW"}
	power_average     = gauge{"tegra.power.average", "Average power consumption", "mW"}
)

// builder adds a metric on its first data point so absent engines leave no empty metric.
type builder struct {
	ms  pmetric.MetricSlice
	ts  pcommon.Timestamp
	dps map[string]pmetric.NumberDataPointSlice
}

func (b *builder) record(g gauge, v int64, attrs ...string) {
	dps, ok := b.dps[g.name]
	if !ok {
		m := b.ms.AppendEmpty()
		m.SetName(g.name)
		m.SetDescription(g.desc)
		m.SetUnit(g.unit)
		dps = m.SetEmptyGauge().DataPoints()
		b.dps[g.name] = dps
	}

	dp := dps.AppendEmpty()
	dp.SetTimestamp(b.ts)
	dp.SetIntValue(v)
	for i := 0; i+1 < len(attrs); i += 2 {
		dp.Attributes().PutStr(attrs[i], attrs[i+1])
	}
}

// toMetrics stamps data points with `ts` rather than `v.Time`:
// tegrastats prints local time without a zone and the parser reads it as UTC.
func toMetrics(v *stats.Stat, ts pcommon.Timestamp) pmetric.Metrics {
	md := pmetric.NewMetrics()
	sm := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
	sm.Scope().SetName(ScopeName)

	b := &builder{
		ms:  sm.Metrics(),
		ts:  ts,
		dps: map[string]pmetric.NumberDataPointSlice{},
	}
	index := func(i int) []string { return []string{"index", strconv.Itoa(i)} }
	sensor := func(name string) []string { return []string{"sensor", name} }

	if !isZero(v.Ram) {
		b.record(ram_inuse, int64(v.Ram.InUse))
		b.record(ram_total, int64(v.Ram.Total))
		b.record(ram_lfb_count, int64(v.Ram.LfbCount))
		b.record(ram_lfb_size, int64(v.Ram.LfbSize))
	}
	if !isZero(v.IRam) {
		b.record(iram_inuse, int64(v.IRam.InUse))
		b.record(iram_total, int64(v.IRam.Total))
	}
	if !isZero(v.Swap) {
		b.record(swap_inuse, int64(v.Swap.InUse))
		b.record(swap_total, int64(v.Swap.Total))
		b.record(swap_cached, int64(v.Swap.Cached))
	}
	for i, w := range v.Cpus {
		if w.Offline {
			continue
		}
		b.record(cpu_utilization, int64(w.Percent), index(i)...)
		b.record(cpu_frequency, int64(w.Freq), index(i)...)
	}
	if !isZero(v.Emc) {
		b.record(emc_utilization, int64(v.Emc.Percent))
		b.record(emc_frequency, int64(v.Emc.Freq))
	}
	if v.Gr3d.Freq != nil {
		b.record(gr3d_utilization, int64(v.Gr3d.Percent))
		for i, w := range v.Gr3d.Freq {
			b.record(gr3d_frequency, int64(w), index(i)...)
		}
	}
	if !isZero(v.Vic) {
		b.record(vic_utilization, int64(v.Vic.Percent))
		b.record(vic_frequency, int64(v.Vic.Freq))
	}
	if !isZero(v.Ape) {
		b.record(ape_frequency, int64(v.Ape.Freq))
	}
	for i, w := range v.NvEnc {
		b.record(nvenc_utilization, int64(w.Percent), index(i)...)
		b.record(nvenc_frequency, int64(w.Freq), index(i)...)
	}
	for i, w := range v.NvDec {
		b.record(nvdec_utilization, int64(w.Percent), index(i)...)
		b.record(nvdec_frequency, int64(w.Freq), index(i)...)
	}
	for i, w := range v.NvDla {
		b.record(nvdla_utilization, int64(w.Percent), index(i)...)
		b.record(nvdla_frequency, int64(w.Freq), index(i)...)
	}
	for i, w := range v.NvJpg {
		b.record(nvjpg_utilization, int64(w.Percent), index(i)...)
		b.record(nvjpg_frequency, int64(w.Freq), index(i)...)
	}
	if !isZero(v.Ofa) {
		b.record(ofa_utilization, int64(v.Ofa.Percent))
		b.record(ofa_frequency, int64(v.Ofa.Freq))
	}
	for _, w := range v.Temp {
		b.record(temperature, int64(w.Value), sensor(w.Name)...)
	}
	for _, w := range v.Power {
		b.record(power_current, int64(w.Current), sensor(w.Name)...)
		b.record(power_average, int64(w.Average), sensor(w.Name)...)
	}

	return md
}

func isZero[T comparable](v T) bool {
	var zero T
	return v == zero
}
