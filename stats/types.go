package stats

import (
	"time"

	"github.com/lesomnus/signals"
)

// Based on:
// - https://docs.nvidia.com/drive/drive_os_5.1.6.1L/nvvib_docs/index.html#page/DRIVE_OS_Linux_SDK_Development_Guide/Utilities/util_tegrastats.html
// - https://docs.nvidia.com/jetson/archives/r38.4/DeveloperGuide/AT/JetsonLinuxDevelopmentTools/TegrastatsUtility.html
// and collected real text example from Jetson AGX Orin Developer Kit.

type Listener = signals.Listener[*Stat]

type Stat struct {
	Time  time.Time
	Ram   Ram
	Swap  Swap
	IRam  IRam
	Cpus  []Cpu
	Emc   Emc
	Gr3d  Gr3d
	Vic   Vic
	Ape   Ape
	NvEnc []NvEnc
	NvDec []NvDec
	NvDla []NvDla
	NvJpg []NvJpg
	Pva   []Pva
	Ofa   Ofa
	Temp  []Temp
	Power []Power
}

// Ram
// Largest Free Block (lfb) is a statistic about the memory allocator.
// It refers to the largest contiguous block of physical memory that can currently be allocated: at most 4 MB.
// It can become smaller with memory fragmentation.
//
//	RAM X/Y (lfb NxZ)
//
//	RAM 18194/62841MB (lfb 466x4MB)
type Ram struct {
	// Amount of RAM in use in megabytes.
	InUse uint // X
	// Total amount of RAM available for applications in megabytes.
	Total uint // Y
	// Number of free blocks of LfbSize.
	LfbCount uint // N
	// Size of the largest free block in megabytes.
	LfbSize uint // Z
}

// Swap
//
//	SWAP X/Y (cached Z)
//
//	SWAP 86/31420MB (cached 0MB)
type Swap struct {
	// Amount of SWAP in use in megabytes.
	InUse uint // X
	// Total amount of SWAP available for applications in megabytes.
	Total uint // Y
	// Amount of SWAP cached in megabytes.
	Cached uint // Z
}

// IRam is memory local to the video hardware engine.
// IRAM stands for Internal RAM, a small on-chip SRAM region integrated directly into the Tegra SoC rather than external system DRAM.
// It is designed for low-latency, high-speed access and is primarily used by hardware components such as multimedia engines, video codecs, ISP, VIC, firmware, and DMA-related operations.
//
//	IRAM X/Y (lfb Z)
type IRam struct {
	// Amount of IRAM memory in use in kilobytes.
	InUse uint // X
	// Total amount of IRAM memory available in kilobytes.
	Total uint // Y
	// Size of the largest free block in kilobytes.
	LfbSize uint // Z
}

// Cpu
//
//	CPU [X%,X%,..]@Z
//	CPU [X%@Z,X%@Z,...]
//
//	CPU [17%@729,3%@729,...,18%@729]
type Cpu struct {
	// Load statistics for each of the CPU cores relative to the current running frequency.
	Percent uint // X
	// CPU frequency in megahertz.
	Freq uint // Z
}

// Emc is External memory controller statistics.
// All sysmem/carve-out/GART memory accesses go through the memory controller.
//
//	EMC_FREQ X%
//	EMC_FREQ @Y
//	EMC_FREQ X%@Y
//
//	EMC_FRQQ 0%@2133
type Emc struct {
	// Percentage of EMC memory bandwidth in use relative to the current running frequency.
	Percent uint // X
	// EMC frequency in megahertz.
	Freq uint // Y
}

// Gr3d is the GPU engine.
// The GPU of Jetson AGX Orin series contains two GPCs and the GPU of Jetson Thor contains three GPCs.
// Each GPC has its own frequency controller.
//
//	GR3D_FREQ X%
//	GR3D_FREQ @[Y, Y, ...]
//	GR3D_FREQ X %@[Y, Y, ...]
//
//	GR3D_FREQ 0%@[0,0]
type Gr3d struct {
	// Proportion of GPU activation time in a period.
	// Different GPCs have the same percentage.
	Percent uint // X
	// Frequency of each of the GPU's GPCs in megahertz.
	Freq []uint // Y
}

// Vic (Video Image Compositor) engine statistics.
// The VIC engine implements video post-processing functions needed by video playback applications to produce a final image for the player window.
//
//	VIC_LOAD X %
//	VIC_FREQ @Y
//	VIC X%@Y
//
//	VIC 0%@729
type Vic struct {
	// VIC engine loading as a percentage of current VIC engine frequency.
	Percent uint // X
	// Current VIC engine frequency.
	Freq uint // Y
}

// Ape is the audio processing engine.
// The APE subsystem consists of ADSP (Cortex®-A9 CPU), mailboxes, AHUB, ADMA, etc.
//
//	APE Y
type Ape struct {
	// APE frequency in megahertz.
	Freq uint // Y
}

// Video hardware encoding engine statistics.
// Shown only when the hardware encoder engine is used.
//
//	NVENC[N] X%
//	NVENC[N] @Y
//	NVENC[N] X%@Y
//
//	NVENC off
//	NVENC 78%@179
type NvEnc struct {
	// NVENC utilization.
	Percent uint // X
	// NVENC frequency in megahertz.
	Freq uint // Y
}

// Video hardware decoding engine statistics.
// Shown only when the hardware decoder engine is used.
//
//	NVDEC[N] X%
//	NVDEC[N] @Y
//	NVDEC[N] X%@Y
//
//	NVDEC off
type NvDec struct {
	// NVDEC utilization.
	Percent uint // X
	// NVDEC frequency in megahertz.
	Freq uint // Y
}

// NvDla is a NVIDIA deep learning accelerator.
// NVDLA1, which is the second instance of NVDLA, might be reported if it has been enabled.
// NVDLA is supported on Jetson AGX Orin but not on Jetson Thor.
//
//	NVDLA[N] X%
//	NVDLA[N] @Y
//	NVDLA[N] X%@Y
type NvDla struct {
	// NVDLA utilization.
	Percent uint // X
	// NVDLA frequency in megahertz.
	Freq uint // Y
}

// NvJpg is a NVIDIA JPEG encoder/decoder.
// NVJPG0 and NVJPG1 are the first and second instances of NVJPG, respectively.
//
//	NVJPG[N] X%
//	NVJPG[N] @Y
//	NVJPG[N] X%@Y
//
//	NVJPG off
//	NVJPG1 off
type NvJpg struct {
	// NVJPG utilization.
	Percent uint // X
	// NVJPG frequency in megahertz.
	Freq uint // Y
}

// Pva is a NVIDIA programmable vision accelerator.
// PVA0_FREQ and PVA1_FREQ are the first and second instances of PVA, respectively.
//
//	PVA[N]_FREQ @Y
//	PVA[N]_FREQ [X%,X%]@Y
//
//	PVA0_FREQ off
type Pva struct {
	// VPU utilizations (VPU is the vector processing unit)
	Percent []uint // X
	// PVA frequency in megahertz.
	Freq uint // Y
}

// Ofa is a NVIDIA optical flow accelerator.
//
//	OFA Y
//	OFA X%@Y
//
//	OFA off
type Ofa struct {
	// OFA utilization.
	Percent uint // X
	// OFA frequency in megahertz.
	Freq uint // Y
}

// Temperature of one of the processor blocks as reported by node `/sys/devices/virtual/thermal/thermal_zone<x>/temp/`, where <x> is the name of the block.
//
//	X@YC
//
//	cpu@52.125C
//	soc2@48.406C
type Temp struct {
	// Processor block name.
	Name string // X
	// Processor block temperature in degrees Celsius
	Value float32 // Y
}

// Power consumption statistics for a power block.
//
//	VDD_X YmW/ZmW
//
//	VDD_IN 3000/3000
//	VDD_CPU_GPU_CV 0/0
//	VDD_SOC 1021/1021
type Power struct {
	// Name of the power rail
	Name string // X
	// Block’s current power consumption in milliwatts.
	Current uint // Y
	// Block’s average power consumption in milliwatts
	Average uint // Z
}
