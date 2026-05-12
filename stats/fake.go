package stats

import (
	"context"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/lesomnus/signals"
)

var _ Listener = (*Fake)(nil)

// Fake generates random Stat values.
type Fake struct {
	event signals.Event[*Stat]

	mu     sync.Mutex
	count  int
	cancel context.CancelFunc
}

func NewFake() *Fake {
	return &Fake{
		event: signals.NewEvent[*Stat](),
	}
}

func (x *Fake) Listen(f signals.Callback[*Stat]) func() {
	stop := x.event.Listen(f)

	x.mu.Lock()
	defer x.mu.Unlock()

	x.count++
	if x.count == 1 {
		ctx, cancel := context.WithCancel(context.Background())
		x.cancel = cancel
		go x.run(ctx)
	}

	return func() {
		stop()
		x.mu.Lock()
		defer x.mu.Unlock()

		x.count--
		if x.count == 0 {
			x.cancel()
		}
	}
}

func (x *Fake) run(ctx context.Context) {
	stat := seedStat()
	x.event.Emit(cloneStat(stat))

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			walkStat(stat)
			x.event.Emit(cloneStat(stat))
		}
	}
}

func walkU(v, lo, hi uint, step uint) uint {
	delta := int(rand.UintN(step + 1))
	if rand.IntN(2) == 0 {
		delta = -delta
	}
	result := int(v) + delta
	if result < int(lo) {
		return lo
	}
	if result > int(hi) {
		return hi
	}
	return uint(result)
}

func walkF32(v, lo, hi, step float32) float32 {
	v += (rand.Float32()*2 - 1) * step
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func cloneStat(s *Stat) *Stat {
	c := *s
	c.Cpus = slices.Clone(s.Cpus)
	c.Gr3d.Freq = slices.Clone(s.Gr3d.Freq)
	c.NvEnc = slices.Clone(s.NvEnc)
	c.Temp = slices.Clone(s.Temp)
	c.Power = slices.Clone(s.Power)
	return &c
}

func seedStat() *Stat {
	cpus := make([]Cpu, 12)
	for i := range cpus {
		cpus[i] = Cpu{
			Percent: uint(rand.Float64() * 40),
			Freq:    uint(729 + rand.Float64()*(2201-729)),
		}
	}

	return &Stat{
		Ram: Ram{
			InUse:    uint(10000 + rand.Float64()*20000),
			Total:    62841,
			LfbCount: 466,
			LfbSize:  4,
		},
		Swap: Swap{
			InUse: uint(rand.Float64() * 200),
			Total: 31420,
		},
		Cpus: cpus,
		Emc: Emc{
			Percent: uint(1 + rand.Float64()*5),
			Freq:    2133,
		},
		Gr3d: Gr3d{
			Percent: uint(rand.Float64() * 30),
			Freq:    []uint{uint(rand.Float64() * 1000), uint(rand.Float64() * 1000)},
		},
		Vic: Vic{
			Percent: uint(20 + rand.Float64()*30),
			Freq:    115,
		},
		Ape: Ape{Freq: 174},
		NvEnc: []NvEnc{
			{
				Percent: uint(50 + rand.Float64()*40),
				Freq:    uint(115 + rand.Float64()*365),
			},
		},
		Temp: []Temp{
			{Name: "cpu", Value: float32(45 + rand.Float64()*15)},
			{Name: "soc2", Value: float32(43 + rand.Float64()*10)},
			{Name: "soc0", Value: float32(43 + rand.Float64()*10)},
			{Name: "tj", Value: float32(45 + rand.Float64()*15)},
			{Name: "soc1", Value: float32(43 + rand.Float64()*10)},
		},
		Power: []Power{
			{Name: "VDD_GPU_SOC", Current: uint(1500 + rand.Float64()*2000)},
			{Name: "VDD_CPU_CV", Current: uint(800 + rand.Float64()*1500)},
			{Name: "VIN_SYS_5V0", Current: uint(3500 + rand.Float64()*2000)},
		},
	}
}

func walkStat(s *Stat) {
	s.Time = time.Now()
	s.Ram.InUse = walkU(s.Ram.InUse, 5000, 55000, 300)
	s.Swap.InUse = walkU(s.Swap.InUse, 0, 500, 15)

	for i := range s.Cpus {
		s.Cpus[i].Percent = walkU(s.Cpus[i].Percent, 0, 100, 8)
		s.Cpus[i].Freq = walkU(s.Cpus[i].Freq, 729, 2201, 150)
	}

	s.Emc.Percent = walkU(s.Emc.Percent, 0, 20, 1)
	s.Gr3d.Percent = walkU(s.Gr3d.Percent, 0, 100, 8)
	for i := range s.Gr3d.Freq {
		s.Gr3d.Freq[i] = walkU(s.Gr3d.Freq[i], 0, 1300, 150)
	}

	for i := range s.NvEnc {
		s.NvEnc[i].Percent = walkU(s.NvEnc[i].Percent, 0, 100, 8)
		s.NvEnc[i].Freq = walkU(s.NvEnc[i].Freq, 115, 480, 25)
	}

	s.Vic.Percent = walkU(s.Vic.Percent, 0, 100, 8)

	for i := range s.Temp {
		s.Temp[i].Value = walkF32(s.Temp[i].Value, 35, 85, 0.5)
	}

	for i := range s.Power {
		s.Power[i].Current = walkU(s.Power[i].Current, 200, 10000, 200)
		s.Power[i].Average = s.Power[i].Current
	}
}
