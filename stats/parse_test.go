package stats_test

import (
	"testing"
	"time"

	"github.com/lesomnus/tegra-exporter/stats"
)

func TestParse(t *testing.T) {
	// Use result of `uname -r` as a title of the title case and please attach result of:
	//
	//	cat /etc/nv_tegra_release
	//	dpkg-query --show nvidia-l4t-core

	type TestCase struct {
		Text string
		Stat *stats.Stat
	}
	validate := func(t *testing.T, tcs []TestCase) {
		t.Helper()
		for i, tc := range tcs {
			t.Run("", func(t *testing.T) {
				err := stats.Parse(tc.Text, tc.Stat)
				if err != nil {
					t.Fatalf("[%d]: unexpected error: %v", i, err)
				}
			})
		}
	}

	// R36 (release), REVISION: 4.4, GCID: 41062509, BOARD: generic, EABI: aarch64, DATE: Mon Jun 16 16:07:13 UTC 2025
	// nvidia-l4t-core 36.4.4-20250616085344
	t.Run("5.15.148", func(t *testing.T) {
		tcs := []TestCase{
			{
				Text: `05-09-2026 18:42:06 RAM 18207/62841MB (lfb 466x4MB) SWAP 86/31420MB (cached 0MB) CPU [14%@2188,5%@2112,8%@2035,14%@2201,46%@1267,47%@1267,2%@1267,10%@1267,4%@1036,5%@1036,6%@1036,10%@1036] EMC_FREQ 3%@2133 GR3D_FREQ 0%@[0,0] NVENC 79%@166 NVDEC off NVJPG off NVJPG1 off VIC 36%@115 OFA off NVDLA0 off NVDLA1 off PVA0_FREQ off APE 174 cpu@52.562C soc2@48.437C soc0@49.093C tj@52.562C soc1@49.656C VDD_GPU_SOC 2412mW/2412mW VDD_CPU_CV 1607mW/1607mW VIN_SYS_5V0 4707mW/4707mW`,
				Stat: &stats.Stat{
					Time: time.Date(2026, 5, 9, 18, 42, 6, 0, time.UTC),
					Ram:  stats.Ram{InUse: 18207, Total: 62841, LfbCount: 466, LfbSize: 4},
					Swap: stats.Swap{InUse: 86, Total: 31420, Cached: 0},
					Cpus: []stats.Cpu{
						{Percent: 14, Freq: 2188},
						{Percent: 5, Freq: 2112},
						{Percent: 8, Freq: 2035},
						{Percent: 14, Freq: 2201},
						{Percent: 46, Freq: 1267},
						{Percent: 47, Freq: 1267},
						{Percent: 2, Freq: 1267},
						{Percent: 10, Freq: 1267},
						{Percent: 4, Freq: 1036},
						{Percent: 5, Freq: 1036},
						{Percent: 6, Freq: 1036},
						{Percent: 10, Freq: 1036},
					},
					Emc:  stats.Emc{Percent: 3, Freq: 2133},
					Gr3d: stats.Gr3d{Percent: 0, Freq: []uint{0, 0}},
					NvEnc: []stats.NvEnc{
						{Percent: 79, Freq: 166},
					},
					Vic: stats.Vic{Percent: 36, Freq: 115},
					Ape: stats.Ape{Freq: 174},
					Temp: []stats.Temp{
						{Name: "cpu", Value: 52.562},
						{Name: "soc2", Value: 48.437},
						{Name: "soc0", Value: 49.093},
						{Name: "tj", Value: 52.562},
						{Name: "soc1", Value: 49.656},
					},
					Power: []stats.Power{
						{Name: "VDD_GPU_SOC", Current: 2412, Average: 2412},
						{Name: "VDD_CPU_CV", Current: 1607, Average: 1607},
						{Name: "VIN_SYS_5V0", Current: 4707, Average: 4707},
					},
				},
			},
			{
				Text: `05-09-2026 18:42:07 RAM 18202/62841MB (lfb 466x4MB) SWAP 86/31420MB (cached 0MB) CPU [16%@1949,6%@1761,4%@1222,4%@2201,4%@1036,33%@1036,0%@1036,40%@1036,3%@1190,46%@1190,6%@1190,2%@1190] EMC_FREQ 3%@2133 GR3D_FREQ 0%@[0,0] NVENC 78%@166 NVDEC off NVJPG off NVJPG1 off VIC 32%@115 OFA off NVDLA0 off NVDLA1 off PVA0_FREQ off APE 174 cpu@52.343C soc2@48.343C soc0@49.062C tj@52.343C soc1@49.531C VDD_GPU_SOC 2412mW/2412mW VDD_CPU_CV 1607mW/1607mW VIN_SYS_5V0 4607mW/4657mW`,
				Stat: &stats.Stat{
					Time: time.Date(2026, 5, 9, 18, 42, 7, 0, time.UTC),
					Ram:  stats.Ram{InUse: 18202, Total: 62841, LfbCount: 466, LfbSize: 4},
					Swap: stats.Swap{InUse: 86, Total: 31420, Cached: 0},
					Cpus: []stats.Cpu{
						{Percent: 16, Freq: 1949},
						{Percent: 6, Freq: 1761},
						{Percent: 4, Freq: 1222},
						{Percent: 4, Freq: 2201},
						{Percent: 4, Freq: 1036},
						{Percent: 33, Freq: 1036},
						{Percent: 0, Freq: 1036},
						{Percent: 40, Freq: 1036},
						{Percent: 3, Freq: 1190},
						{Percent: 46, Freq: 1190},
						{Percent: 6, Freq: 1190},
						{Percent: 2, Freq: 1190},
					},
					Emc:  stats.Emc{Percent: 3, Freq: 2133},
					Gr3d: stats.Gr3d{Percent: 0, Freq: []uint{0, 0}},
					NvEnc: []stats.NvEnc{
						{Percent: 78, Freq: 166},
					},
					Vic: stats.Vic{Percent: 32, Freq: 115},
					Ape: stats.Ape{Freq: 174},
					Temp: []stats.Temp{
						{Name: "cpu", Value: 52.343},
						{Name: "soc2", Value: 48.343},
						{Name: "soc0", Value: 49.062},
						{Name: "tj", Value: 52.343},
						{Name: "soc1", Value: 49.531},
					},
					Power: []stats.Power{
						{Name: "VDD_GPU_SOC", Current: 2412, Average: 2412},
						{Name: "VDD_CPU_CV", Current: 1607, Average: 1607},
						{Name: "VIN_SYS_5V0", Current: 4607, Average: 4657},
					},
				},
			},
			{
				Text: `05-09-2026 18:42:08 RAM 18203/62841MB (lfb 466x4MB) SWAP 86/31420MB (cached 0MB) CPU [14%@729,5%@1648,4%@1645,3%@2050,3%@1113,0%@1113,0%@1113,70%@1113,4%@1651,32%@1651,3%@1651,15%@1651] EMC_FREQ 2%@2133 GR3D_FREQ 0%@[0,0] NVENC 78%@179 NVDEC off NVJPG off NVJPG1 off VIC 37%@115 OFA off NVDLA0 off NVDLA1 off PVA0_FREQ off APE 174 cpu@52.156C soc2@48.468C soc0@49.218C tj@52.156C soc1@49.625C VDD_GPU_SOC 2412mW/2412mW VDD_CPU_CV 1607mW/1607mW VIN_SYS_5V0 4707mW/4674mW`,
				Stat: &stats.Stat{
					Time: time.Date(2026, 5, 9, 18, 42, 8, 0, time.UTC),
					Ram:  stats.Ram{InUse: 18203, Total: 62841, LfbCount: 466, LfbSize: 4},
					Swap: stats.Swap{InUse: 86, Total: 31420, Cached: 0},
					Cpus: []stats.Cpu{
						{Percent: 14, Freq: 729},
						{Percent: 5, Freq: 1648},
						{Percent: 4, Freq: 1645},
						{Percent: 3, Freq: 2050},
						{Percent: 3, Freq: 1113},
						{Percent: 0, Freq: 1113},
						{Percent: 0, Freq: 1113},
						{Percent: 70, Freq: 1113},
						{Percent: 4, Freq: 1651},
						{Percent: 32, Freq: 1651},
						{Percent: 3, Freq: 1651},
						{Percent: 15, Freq: 1651},
					},
					Emc:  stats.Emc{Percent: 2, Freq: 2133},
					Gr3d: stats.Gr3d{Percent: 0, Freq: []uint{0, 0}},
					NvEnc: []stats.NvEnc{
						{Percent: 78, Freq: 179},
					},
					Vic: stats.Vic{Percent: 37, Freq: 115},
					Ape: stats.Ape{Freq: 174},
					Temp: []stats.Temp{
						{Name: "cpu", Value: 52.156},
						{Name: "soc2", Value: 48.468},
						{Name: "soc0", Value: 49.218},
						{Name: "tj", Value: 52.156},
						{Name: "soc1", Value: 49.625},
					},
					Power: []stats.Power{
						{Name: "VDD_GPU_SOC", Current: 2412, Average: 2412},
						{Name: "VDD_CPU_CV", Current: 1607, Average: 1607},
						{Name: "VIN_SYS_5V0", Current: 4707, Average: 4674},
					},
				},
			},
		}
		validate(t, tcs)
	})
}
