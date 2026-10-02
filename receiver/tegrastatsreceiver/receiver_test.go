package tegrastatsreceiver_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/lesomnus/tegra-exporter/receiver/tegrastatsreceiver"
	"github.com/lesomnus/tegra-exporter/stats"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

// From Jetson AGX Orin, R36.4.4.
const line = `05-09-2026 18:42:06 RAM 18207/62841MB (lfb 466x4MB) SWAP 86/31420MB (cached 0MB) CPU [14%@2188,5%@2112,8%@2035,14%@2201,46%@1267,47%@1267,2%@1267,10%@1267,4%@1036,5%@1036,6%@1036,10%@1036] EMC_FREQ 3%@2133 GR3D_FREQ 0%@[0,0] NVENC 79%@166 NVDEC off NVJPG off NVJPG1 off VIC 36%@115 OFA off NVDLA0 off NVDLA1 off PVA0_FREQ off APE 174 cpu@52.562C soc2@48.437C soc0@49.093C tj@52.562C soc1@49.656C VDD_GPU_SOC 2412mW/2412mW VDD_CPU_CV 1607mW/1607mW VIN_SYS_5V0 4707mW/4707mW`

// Prints `line` repeatedly, standing in for `tegrastats`.
var command = []string{"sh", "-c", fmt.Sprintf("while :; do echo '%s'; sleep 0.05; done", line)}

func TestConfigValidate(t *testing.T) {
	f := tegrastatsreceiver.NewFactory()
	if err := componenttest.CheckConfigStruct(f.CreateDefaultConfig()); err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		desc  string
		edit  func(c *tegrastatsreceiver.Config)
		valid bool
	}{
		{"default", func(c *tegrastatsreceiver.Config) {}, true},
		{"scrape", func(c *tegrastatsreceiver.Config) { c.Mode = tegrastatsreceiver.ModeScrape }, true},
		{"unknown mode", func(c *tegrastatsreceiver.Config) { c.Mode = "pull" }, false},
		{"empty command", func(c *tegrastatsreceiver.Config) { c.Command = nil }, false},
		{"blank command", func(c *tegrastatsreceiver.Config) { c.Command = []string{""} }, false},
		{"scrape without stale timeout", func(c *tegrastatsreceiver.Config) {
			c.Mode = tegrastatsreceiver.ModeScrape
			c.StaleTimeout = 0
		}, false},
		{"push ignores stale timeout", func(c *tegrastatsreceiver.Config) { c.StaleTimeout = 0 }, true},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			c := f.CreateDefaultConfig().(*tegrastatsreceiver.Config)
			tc.edit(c)
			err := c.Validate()
			if tc.valid && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestReceiver(t *testing.T) {
	tcs := []struct {
		desc string
		edit func(c *tegrastatsreceiver.Config)
	}{
		{"push", func(c *tegrastatsreceiver.Config) {
			c.Mode = tegrastatsreceiver.ModePush
		}},
		{"scrape", func(c *tegrastatsreceiver.Config) {
			c.Mode = tegrastatsreceiver.ModeScrape
			c.CollectionInterval = 50 * time.Millisecond
			c.InitialDelay = 0
		}},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			f := tegrastatsreceiver.NewFactory()
			c := f.CreateDefaultConfig().(*tegrastatsreceiver.Config)
			c.Command = command
			tc.edit(c)

			sink := new(consumertest.MetricsSink)
			r, err := f.CreateMetrics(t.Context(), receivertest.NewNopSettings(tegrastatsreceiver.Type), c, sink)
			if err != nil {
				t.Fatal(err)
			}
			if err := r.Start(t.Context(), componenttest.NewNopHost()); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := r.Shutdown(t.Context()); err != nil {
					t.Fatal(err)
				}
			}()

			// Scrapes before the first line arrives come out empty.
			deadline := time.Now().Add(5 * time.Second)
			for !hasData(sink) {
				if time.Now().After(deadline) {
					t.Fatal("no metrics received")
				}
				time.Sleep(10 * time.Millisecond)
			}

			for _, md := range sink.AllMetrics() {
				if md.DataPointCount() == 0 {
					continue
				}
				assertCpu0(t, md)
				return
			}
		})
	}
}

func TestFake(t *testing.T) {
	f := tegrastatsreceiver.NewFactory()
	c := f.CreateDefaultConfig().(*tegrastatsreceiver.Config)
	c.Command = []string{"$fake"}

	sink := new(consumertest.MetricsSink)
	r, err := f.CreateMetrics(t.Context(), receivertest.NewNopSettings(tegrastatsreceiver.Type), c, sink)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Start(t.Context(), componenttest.NewNopHost()); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for !hasData(sink) {
		if time.Now().After(deadline) {
			t.Fatal("no metrics received")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := r.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestShutdownWithoutStart(t *testing.T) {
	f := tegrastatsreceiver.NewFactory()
	for _, mode := range []tegrastatsreceiver.Mode{tegrastatsreceiver.ModePush, tegrastatsreceiver.ModeScrape} {
		c := f.CreateDefaultConfig().(*tegrastatsreceiver.Config)
		c.Mode = mode
		r, err := f.CreateMetrics(t.Context(), receivertest.NewNopSettings(tegrastatsreceiver.Type), c, consumertest.NewNop())
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Shutdown(t.Context()); err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
	}
}

func TestMetrics(t *testing.T) {
	var v stats.Stat
	if err := stats.Parse(line, &v); err != nil {
		t.Fatal(err)
	}
	// Exercises values the line above does not have.
	v.Cpus[1].Offline = true
	v.Temp = append(v.Temp, stats.Temp{Name: "cv0", Value: -256})

	md := tegrastatsreceiver.ToMetrics(&v, pcommon.NewTimestampFromTime(time.Now()))
	got := map[string]map[string]int64{}
	ms := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	for i := range ms.Len() {
		m := ms.At(i)
		if _, ok := got[m.Name()]; ok {
			t.Fatalf("duplicated metric: %s", m.Name())
		}
		if m.Type() != pmetric.MetricTypeGauge {
			t.Fatalf("%s: not a gauge", m.Name())
		}

		vs := map[string]int64{}
		dps := m.Gauge().DataPoints()
		for j := range dps.Len() {
			dp := dps.At(j)
			key := ""
			dp.Attributes().Range(func(k string, a pcommon.Value) bool {
				key = k + "=" + a.Str()
				return true
			})
			vs[key] = dp.IntValue()
		}
		got[m.Name()] = vs
	}

	want := map[string]map[string]int64{
		"tegra.ram.in_use":        {"": 18207},
		"tegra.ram.total":         {"": 62841},
		"tegra.ram.lfb_count":     {"": 466},
		"tegra.ram.lfb_size":      {"": 4},
		"tegra.swap.in_use":       {"": 86},
		"tegra.swap.total":        {"": 31420},
		"tegra.swap.cached":       {"": 0},
		"tegra.emc.utilization":   {"": 3},
		"tegra.emc.frequency":     {"": 2133},
		"tegra.gr3d.utilization":  {"": 0},
		"tegra.gr3d.frequency":    {"index=0": 0, "index=1": 0},
		"tegra.vic.utilization":   {"": 36},
		"tegra.vic.frequency":     {"": 115},
		"tegra.ape.frequency":     {"": 174},
		"tegra.nvenc.utilization": {"index=0": 79},
		"tegra.nvenc.frequency":   {"index=0": 166},
		"tegra.temperature": {
			"sensor=cpu": 52, "sensor=soc2": 48, "sensor=soc0": 49, "sensor=tj": 52, "sensor=soc1": 49,
			"sensor=cv0": -256,
		},
		"tegra.power.current": {"sensor=VDD_GPU_SOC": 2412, "sensor=VDD_CPU_CV": 1607, "sensor=VIN_SYS_5V0": 4707},
		"tegra.power.average": {"sensor=VDD_GPU_SOC": 2412, "sensor=VDD_CPU_CV": 1607, "sensor=VIN_SYS_5V0": 4707},
	}
	for _, name := range []string{"tegra.cpu.utilization", "tegra.cpu.frequency"} {
		got_cpu := got[name]
		delete(got, name)
		if len(got_cpu) != 11 {
			t.Fatalf("%s: want 11 cores without the offline one, got %d", name, len(got_cpu))
		}
		if _, ok := got_cpu["index=1"]; ok {
			t.Fatalf("%s: offline core is reported", name)
		}
	}

	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("missing %s", name)
			continue
		}
		delete(got, name)
		for k, wv := range w {
			if gv, ok := g[k]; !ok || gv != wv {
				t.Errorf("%s{%s}: got %d (%v), want %d", name, k, gv, ok, wv)
			}
		}
		if len(g) != len(w) {
			t.Errorf("%s: got %d points, want %d", name, len(g), len(w))
		}
	}
	for name := range got {
		t.Errorf("unexpected %s", name)
	}
}

func hasData(sink *consumertest.MetricsSink) bool {
	for _, md := range sink.AllMetrics() {
		if md.DataPointCount() > 0 {
			return true
		}
	}
	return false
}

func assertCpu0(t *testing.T, md pmetric.Metrics) {
	t.Helper()
	ms := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	for i := range ms.Len() {
		m := ms.At(i)
		if m.Name() != "tegra.cpu.frequency" {
			continue
		}
		dp := m.Gauge().DataPoints().At(0)
		if v, _ := dp.Attributes().Get("index"); v.Str() != "0" || dp.IntValue() != 2188 {
			t.Fatalf("unexpected cpu0 frequency: %v %d", dp.Attributes().AsRaw(), dp.IntValue())
		}
		return
	}
	t.Fatal("tegra.cpu.frequency is missing")
}
