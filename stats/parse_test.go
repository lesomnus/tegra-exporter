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

	// R38 (release), REVISION: 4.0, GCID: 43443517, BOARD: generic, EABI: aarch64, DATE: Wed Dec 31 00:15:19 UTC 2025
	// nvidia-l4t-core	38.4.0-20251230160601
	t.Run("6.8.12", func(t *testing.T) {
		tcs := []TestCase{
			{
				Text: `05-17-2026 10:00:32 RAM 4208/125773MB (lfb 31x4MB) SWAP 0/2048MB (cached 0MB) CPU [1%@972,0%@972,0%@972,0%@972,1%@972,0%@972,1%@972,1%@972,0%@972,0%@972,3%@972,0%@972,0%@972,1%@972] EMC_FREQ 0%@665 GR3D_FREQ @[0,0,0] NVENC0_FREQ @0 NVENC1_FREQ @0 NVDEC0_FREQ @0 NVDEC1_FREQ @0 NVJPG0_FREQ @0 VIC off OFA_FREQ @0 PVA0_FREQ off APE 300 cpu@34.468C tj@34.5C soc012@33.781C soc345@34.5C VDD_GPU 0mW/0mW VDD_CPU_SOC_MSS 4165mW/4165mW VIN_SYS_5V0 4126mW/4126mW`,
				Stat: &stats.Stat{
					Time: time.Date(2026, 5, 17, 10, 0, 32, 0, time.UTC),
					Ram:  stats.Ram{InUse: 4208, Total: 125773, LfbCount: 31, LfbSize: 4},
					Swap: stats.Swap{InUse: 0, Total: 2048, Cached: 0},
					Cpus: []stats.Cpu{
						{Percent: 1, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 1, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 1, Freq: 972},
						{Percent: 1, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 3, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 1, Freq: 972},
					},
					Emc:   stats.Emc{Percent: 0, Freq: 665},
					Gr3d:  stats.Gr3d{Percent: 0, Freq: []uint{0, 0, 0}},
					NvEnc: []stats.NvEnc{{Percent: 0, Freq: 0}, {Percent: 0, Freq: 0}},
					NvDec: []stats.NvDec{{Percent: 0, Freq: 0}, {Percent: 0, Freq: 0}},
					NvJpg: []stats.NvJpg{{Percent: 0, Freq: 0}},
					Ofa:   stats.Ofa{Percent: 0, Freq: 0},
					Ape:   stats.Ape{Freq: 300},
					Temp: []stats.Temp{
						{Name: "cpu", Value: 34.468},
						{Name: "tj", Value: 34.5},
						{Name: "soc012", Value: 33.781},
						{Name: "soc345", Value: 34.5},
					},
					Power: []stats.Power{
						{Name: "VDD_GPU", Current: 0, Average: 0},
						{Name: "VDD_CPU_SOC_MSS", Current: 4165, Average: 4165},
						{Name: "VIN_SYS_5V0", Current: 4126, Average: 4126},
					},
				},
			},
			{
				Text: `05-17-2026 10:00:33 RAM 4280/125773MB (lfb 29x4MB) SWAP 0/2048MB (cached 0MB) CPU [0%@972,0%@972,0%@972,0%@972,1%@972,0%@972,1%@972,0%@972,0%@2601,17%@2601,1%@972,9%@972,0%@972,1%@972] EMC_FREQ 0%@665 GR3D_FREQ @[0,0,0] NVENC0_FREQ @0 NVENC1_FREQ @0 NVDEC0_FREQ @0 NVDEC1_FREQ @0 NVJPG0_FREQ @0 VIC off OFA_FREQ @0 PVA0_FREQ off APE 300 cpu@35.25C tj@36.031C soc012@36.062C soc345@34.5C VDD_GPU 0mW/0mW VDD_CPU_SOC_MSS 4923mW/4544mW VIN_SYS_5V0 4327mW/4227mW`,
				Stat: &stats.Stat{
					Time: time.Date(2026, 5, 17, 10, 0, 33, 0, time.UTC),
					Ram:  stats.Ram{InUse: 4280, Total: 125773, LfbCount: 29, LfbSize: 4},
					Swap: stats.Swap{InUse: 0, Total: 2048, Cached: 0},
					Cpus: []stats.Cpu{
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 1, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 1, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 0, Freq: 2601},
						{Percent: 17, Freq: 2601},
						{Percent: 1, Freq: 972},
						{Percent: 9, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 1, Freq: 972},
					},
					Emc:   stats.Emc{Percent: 0, Freq: 665},
					Gr3d:  stats.Gr3d{Percent: 0, Freq: []uint{0, 0, 0}},
					NvEnc: []stats.NvEnc{{Percent: 0, Freq: 0}, {Percent: 0, Freq: 0}},
					NvDec: []stats.NvDec{{Percent: 0, Freq: 0}, {Percent: 0, Freq: 0}},
					NvJpg: []stats.NvJpg{{Percent: 0, Freq: 0}},
					Ofa:   stats.Ofa{Percent: 0, Freq: 0},
					Ape:   stats.Ape{Freq: 300},
					Temp: []stats.Temp{
						{Name: "cpu", Value: 35.25},
						{Name: "tj", Value: 36.031},
						{Name: "soc012", Value: 36.062},
						{Name: "soc345", Value: 34.5},
					},
					Power: []stats.Power{
						{Name: "VDD_GPU", Current: 0, Average: 0},
						{Name: "VDD_CPU_SOC_MSS", Current: 4923, Average: 4544},
						{Name: "VIN_SYS_5V0", Current: 4327, Average: 4227},
					},
				},
			},
			{
				Text: `05-17-2026 10:00:34 RAM 4309/125773MB (lfb 23x4MB) SWAP 0/2048MB (cached 0MB) CPU [9%@972,2%@972,1%@972,0%@972,2%@972,11%@972,9%@972,17%@972,1%@972,20%@972,2%@972,2%@972,16%@972,3%@972] EMC_FREQ 1%@665 GR3D_FREQ @[0,0,0] NVENC0_FREQ @0 NVENC1_FREQ @0 NVDEC0_FREQ @0 NVDEC1_FREQ @0 NVJPG0_FREQ @0 VIC off OFA_FREQ @0 PVA0_FREQ off APE 300 cpu@34.468C tj@34.468C soc012@33.843C soc345@34.468C VDD_GPU 0mW/0mW VDD_CPU_SOC_MSS 4544mW/4544mW VIN_SYS_5V0 4428mW/4294mW`,
				Stat: &stats.Stat{
					Time: time.Date(2026, 5, 17, 10, 0, 34, 0, time.UTC),
					Ram:  stats.Ram{InUse: 4309, Total: 125773, LfbCount: 23, LfbSize: 4},
					Swap: stats.Swap{InUse: 0, Total: 2048, Cached: 0},
					Cpus: []stats.Cpu{
						{Percent: 9, Freq: 972},
						{Percent: 2, Freq: 972},
						{Percent: 1, Freq: 972},
						{Percent: 0, Freq: 972},
						{Percent: 2, Freq: 972},
						{Percent: 11, Freq: 972},
						{Percent: 9, Freq: 972},
						{Percent: 17, Freq: 972},
						{Percent: 1, Freq: 972},
						{Percent: 20, Freq: 972},
						{Percent: 2, Freq: 972},
						{Percent: 2, Freq: 972},
						{Percent: 16, Freq: 972},
						{Percent: 3, Freq: 972},
					},
					Emc:   stats.Emc{Percent: 1, Freq: 665},
					Gr3d:  stats.Gr3d{Percent: 0, Freq: []uint{0, 0, 0}},
					NvEnc: []stats.NvEnc{{Percent: 0, Freq: 0}, {Percent: 0, Freq: 0}},
					NvDec: []stats.NvDec{{Percent: 0, Freq: 0}, {Percent: 0, Freq: 0}},
					NvJpg: []stats.NvJpg{{Percent: 0, Freq: 0}},
					Ofa:   stats.Ofa{Percent: 0, Freq: 0},
					Ape:   stats.Ape{Freq: 300},
					Temp: []stats.Temp{
						{Name: "cpu", Value: 34.468},
						{Name: "tj", Value: 34.468},
						{Name: "soc012", Value: 33.843},
						{Name: "soc345", Value: 34.468},
					},
					Power: []stats.Power{
						{Name: "VDD_GPU", Current: 0, Average: 0},
						{Name: "VDD_CPU_SOC_MSS", Current: 4544, Average: 4544},
						{Name: "VIN_SYS_5V0", Current: 4428, Average: 4294},
					},
				},
			},
		}
		validate(t, tcs)
	})
}
