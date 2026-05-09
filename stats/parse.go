package stats

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type scanner struct {
	sc       *bufio.Scanner
	peeked   string
	has_peek bool
}

func newScanner(text string) *scanner {
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Split(bufio.ScanWords)
	return &scanner{sc: sc}
}

func (s *scanner) next() (string, bool) {
	if s.has_peek {
		tok := s.peeked
		s.has_peek = false
		return tok, true
	}
	if s.sc.Scan() {
		return s.sc.Text(), true
	}
	return "", false
}

func (s *scanner) peek() (string, bool) {
	if !s.has_peek {
		if s.sc.Scan() {
			s.peeked = s.sc.Text()
			s.has_peek = true
		}
	}
	return s.peeked, s.has_peek
}

func (s *scanner) expect(ctx string) (string, error) {
	tok, ok := s.next()
	if !ok {
		return "", fmt.Errorf("%s: unexpected end", ctx)
	}
	return tok, nil
}

func Parse(text string, stat *Stat) error {
	*stat = Stat{}
	sc := newScanner(text)

	t, err := parseDateTime(sc)
	if err != nil {
		return fmt.Errorf("parse time: %w", err)
	}
	stat.Time = t

	for {
		tok, ok := sc.next()
		if !ok {
			break
		}

		switch {
		case tok == "RAM":
			ram, err := parseRam(sc)
			if err != nil {
				return fmt.Errorf("RAM: %w", err)
			}
			stat.Ram = ram

		case tok == "SWAP":
			swap, err := parseSwap(sc)
			if err != nil {
				return fmt.Errorf("SWAP: %w", err)
			}
			stat.Swap = swap

		case tok == "IRAM":
			iram, err := parseIram(sc)
			if err != nil {
				return fmt.Errorf("IRAM: %w", err)
			}
			stat.IRam = iram

		case tok == "CPU":
			cpus, err := parseCpu(sc)
			if err != nil {
				return fmt.Errorf("CPU: %w", err)
			}
			stat.Cpus = cpus

		case tok == "EMC_FREQ":
			pct, freq, off, err := parsePercentFreq(sc)
			if err != nil {
				return fmt.Errorf("EMC_FREQ: %w", err)
			}
			if !off {
				stat.Emc = Emc{Percent: pct, Freq: freq}
			}

		case tok == "GR3D_FREQ":
			gr3d, off, err := parseGr3d(sc)
			if err != nil {
				return fmt.Errorf("GR3D_FREQ: %w", err)
			}
			if !off {
				stat.Gr3d = gr3d
			}

		case tok == "VIC":
			pct, freq, off, err := parsePercentFreq(sc)
			if err != nil {
				return fmt.Errorf("VIC: %w", err)
			}
			if !off {
				stat.Vic = Vic{Percent: pct, Freq: freq}
			}

		case tok == "APE":
			val, err := sc.expect("APE")
			if err != nil {
				return err
			}
			freq, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("APE: %w", err)
			}
			stat.Ape = Ape{Freq: uint(freq)}

		case tok == "OFA":
			pct, freq, off, err := parsePercentFreq(sc)
			if err != nil {
				return fmt.Errorf("OFA: %w", err)
			}
			if !off {
				stat.Ofa = Ofa{Percent: pct, Freq: freq}
			}

		case strings.HasPrefix(tok, "NVENC"):
			pct, freq, off, err := parsePercentFreq(sc)
			if err != nil {
				return fmt.Errorf("%s: %w", tok, err)
			}
			if !off {
				stat.NvEnc = append(stat.NvEnc, NvEnc{Percent: pct, Freq: freq})
			}

		case strings.HasPrefix(tok, "NVDEC"):
			pct, freq, off, err := parsePercentFreq(sc)
			if err != nil {
				return fmt.Errorf("%s: %w", tok, err)
			}
			if !off {
				stat.NvDec = append(stat.NvDec, NvDec{Percent: pct, Freq: freq})
			}

		case strings.HasPrefix(tok, "NVJPG"):
			pct, freq, off, err := parsePercentFreq(sc)
			if err != nil {
				return fmt.Errorf("%s: %w", tok, err)
			}
			if !off {
				stat.NvJpg = append(stat.NvJpg, NvJpg{Percent: pct, Freq: freq})
			}

		case strings.HasPrefix(tok, "NVDLA"):
			pct, freq, off, err := parsePercentFreq(sc)
			if err != nil {
				return fmt.Errorf("%s: %w", tok, err)
			}
			if !off {
				stat.NvDla = append(stat.NvDla, NvDla{Percent: pct, Freq: freq})
			}

		case strings.HasPrefix(tok, "PVA") && strings.HasSuffix(tok, "_FREQ"):
			pva, off, err := parsePva(sc)
			if err != nil {
				return fmt.Errorf("%s: %w", tok, err)
			}
			if !off {
				stat.Pva = append(stat.Pva, pva)
			}

		default:
			// Temperature: "name@valueC" e.g. "cpu@52.562C"
			if at := strings.IndexByte(tok, '@'); at > 0 && tok[len(tok)-1] == 'C' {
				value, err := strconv.ParseFloat(tok[at+1:len(tok)-1], 32)
				if err == nil {
					stat.Temp = append(stat.Temp, Temp{Name: tok[:at], Value: float32(value)})
					continue
				}
			}
			// Power: "NAME YmW/ZmW"
			if next, ok := sc.peek(); ok {
				slash := strings.IndexByte(next, '/')
				if slash > 0 {
					cur_m := strings.IndexByte(next[:slash], 'm')
					if cur_m > 0 && next[cur_m:slash] == "mW" && len(next) > 2 && next[len(next)-2:] == "mW" {
						current, err1 := strconv.ParseUint(next[:cur_m], 10, 64)
						average, err2 := strconv.ParseUint(next[slash+1:len(next)-2], 10, 64)
						if err1 == nil && err2 == nil {
							sc.next() // consume the value token
							stat.Power = append(stat.Power, Power{Name: tok, Current: uint(current), Average: uint(average)})
						}
					}
				}
			}
		}
	}
	return nil
}

func parseDateTime(sc *scanner) (time.Time, error) {
	date_str, ok := sc.next()
	if !ok {
		return time.Time{}, fmt.Errorf("unexpected end")
	}
	time_str, ok := sc.next()
	if !ok {
		return time.Time{}, fmt.Errorf("unexpected end")
	}
	if len(date_str) != 10 || date_str[2] != '-' || date_str[5] != '-' {
		return time.Time{}, fmt.Errorf("invalid date: %s", date_str)
	}
	if len(time_str) != 8 || time_str[2] != ':' || time_str[5] != ':' {
		return time.Time{}, fmt.Errorf("invalid time: %s", time_str)
	}
	month, err := strconv.Atoi(date_str[0:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("month: %w", err)
	}
	day, err := strconv.Atoi(date_str[3:5])
	if err != nil {
		return time.Time{}, fmt.Errorf("day: %w", err)
	}
	year, err := strconv.Atoi(date_str[6:10])
	if err != nil {
		return time.Time{}, fmt.Errorf("year: %w", err)
	}
	hour, err := strconv.Atoi(time_str[0:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("hour: %w", err)
	}
	minute, err := strconv.Atoi(time_str[3:5])
	if err != nil {
		return time.Time{}, fmt.Errorf("minute: %w", err)
	}
	second, err := strconv.Atoi(time_str[6:8])
	if err != nil {
		return time.Time{}, fmt.Errorf("second: %w", err)
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}

func parseRam(sc *scanner) (Ram, error) {
	size_str, ok := sc.next()
	if !ok {
		return Ram{}, fmt.Errorf("unexpected end")
	}
	sc.next() // "(lfb"
	lfb_str, ok := sc.next()
	if !ok {
		return Ram{}, fmt.Errorf("unexpected end")
	}

	slash := strings.IndexByte(size_str, '/')
	m := strings.IndexByte(size_str, 'M')
	if slash < 0 || m <= slash {
		return Ram{}, fmt.Errorf("invalid size: %s", size_str)
	}
	in_use, err := strconv.ParseUint(size_str[:slash], 10, 64)
	if err != nil {
		return Ram{}, fmt.Errorf("in use: %w", err)
	}
	total, err := strconv.ParseUint(size_str[slash+1:m], 10, 64)
	if err != nil {
		return Ram{}, fmt.Errorf("total: %w", err)
	}
	x := strings.IndexByte(lfb_str, 'x')
	lfb_m := strings.IndexByte(lfb_str, 'M')
	if x < 0 || lfb_m <= x {
		return Ram{}, fmt.Errorf("invalid lfb: %s", lfb_str)
	}
	lfb_count, err := strconv.ParseUint(lfb_str[:x], 10, 64)
	if err != nil {
		return Ram{}, fmt.Errorf("lfb count: %w", err)
	}
	lfb_size, err := strconv.ParseUint(lfb_str[x+1:lfb_m], 10, 64)
	if err != nil {
		return Ram{}, fmt.Errorf("lfb size: %w", err)
	}
	return Ram{InUse: uint(in_use), Total: uint(total), LfbCount: uint(lfb_count), LfbSize: uint(lfb_size)}, nil
}

func parseSwap(sc *scanner) (Swap, error) {
	size_str, ok := sc.next()
	if !ok {
		return Swap{}, fmt.Errorf("unexpected end")
	}
	sc.next() // "(cached"
	cached_str, ok := sc.next()
	if !ok {
		return Swap{}, fmt.Errorf("unexpected end")
	}

	slash := strings.IndexByte(size_str, '/')
	m := strings.IndexByte(size_str, 'M')
	if slash < 0 || m <= slash {
		return Swap{}, fmt.Errorf("invalid size: %s", size_str)
	}
	in_use, err := strconv.ParseUint(size_str[:slash], 10, 64)
	if err != nil {
		return Swap{}, fmt.Errorf("in use: %w", err)
	}
	total, err := strconv.ParseUint(size_str[slash+1:m], 10, 64)
	if err != nil {
		return Swap{}, fmt.Errorf("total: %w", err)
	}
	cached_num, _, ok := strings.Cut(cached_str, "M")
	if !ok {
		return Swap{}, fmt.Errorf("invalid cached: %s", cached_str)
	}
	cached, err := strconv.ParseUint(cached_num, 10, 64)
	if err != nil {
		return Swap{}, fmt.Errorf("cached: %w", err)
	}
	return Swap{InUse: uint(in_use), Total: uint(total), Cached: uint(cached)}, nil
}

func parseIram(sc *scanner) (IRam, error) {
	size_str, ok := sc.next()
	if !ok {
		return IRam{}, fmt.Errorf("unexpected end")
	}
	sc.next() // "(lfb"
	lfb_str, ok := sc.next()
	if !ok {
		return IRam{}, fmt.Errorf("unexpected end")
	}

	slash := strings.IndexByte(size_str, '/')
	k := strings.IndexByte(size_str, 'k')
	if slash < 0 || k <= slash {
		return IRam{}, fmt.Errorf("invalid size: %s", size_str)
	}
	in_use, err := strconv.ParseUint(size_str[:slash], 10, 64)
	if err != nil {
		return IRam{}, fmt.Errorf("in use: %w", err)
	}
	total, err := strconv.ParseUint(size_str[slash+1:k], 10, 64)
	if err != nil {
		return IRam{}, fmt.Errorf("total: %w", err)
	}
	lfb_num, _, ok := strings.Cut(lfb_str, "k")
	if !ok {
		return IRam{}, fmt.Errorf("invalid lfb: %s", lfb_str)
	}
	lfb_size, err := strconv.ParseUint(lfb_num, 10, 64)
	if err != nil {
		return IRam{}, fmt.Errorf("lfb size: %w", err)
	}
	return IRam{InUse: uint(in_use), Total: uint(total), LfbSize: uint(lfb_size)}, nil
}

func parseCpu(sc *scanner) ([]Cpu, error) {
	s, ok := sc.next()
	if !ok {
		return nil, fmt.Errorf("unexpected end")
	}

	if len(s) < 2 || s[0] != '[' {
		return nil, fmt.Errorf("missing '['")
	}
	s = s[1:]

	inner, suffix, ok := strings.Cut(s, "]")
	if !ok {
		return nil, fmt.Errorf("missing ']'")
	}

	n := strings.Count(inner, ",") + 1
	cpus := make([]Cpu, 0, n)

	if len(suffix) > 0 && suffix[0] == '@' {
		freq, err := strconv.ParseUint(suffix[1:], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("freq: %w", err)
		}
		for len(inner) > 0 {
			var part string
			part, inner, _ = strings.Cut(inner, ",")
			pct_end := len(part)
			if pct_end > 0 && part[pct_end-1] == '%' {
				pct_end--
			}
			v, err := strconv.ParseUint(part[:pct_end], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("percent: %w", err)
			}
			cpus = append(cpus, Cpu{Percent: uint(v), Freq: uint(freq)})
		}
	} else {
		for len(inner) > 0 {
			var part string
			part, inner, _ = strings.Cut(inner, ",")
			pct_str, freq_str, ok := strings.Cut(part, "@")
			if !ok {
				return nil, fmt.Errorf("invalid core: %s", part)
			}
			if len(pct_str) > 0 && pct_str[len(pct_str)-1] == '%' {
				pct_str = pct_str[:len(pct_str)-1]
			}
			pct, err := strconv.ParseUint(pct_str, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("percent: %w", err)
			}
			freq, err := strconv.ParseUint(freq_str, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("freq: %w", err)
			}
			cpus = append(cpus, Cpu{Percent: uint(pct), Freq: uint(freq)})
		}
	}
	return cpus, nil
}

// parsePercentFreq reads one token and parses "X%@Y", "@Y", or "X%".
// Returns off=true if the token is "off".
func parsePercentFreq(sc *scanner) (pct, freq uint, off bool, err error) {
	s, ok := sc.next()
	if !ok {
		err = fmt.Errorf("unexpected end")
		return
	}
	if s == "off" {
		off = true
		return
	}

	at := strings.IndexByte(s, '@')
	if at >= 0 {
		if at > 0 {
			pct_end := at
			if s[at-1] == '%' {
				pct_end--
			}
			v, e := strconv.ParseUint(s[:pct_end], 10, 64)
			if e != nil {
				err = fmt.Errorf("percent: %w", e)
				return
			}
			pct = uint(v)
		}
		v, e := strconv.ParseUint(s[at+1:], 10, 64)
		if e != nil {
			err = fmt.Errorf("freq: %w", e)
			return
		}
		freq = uint(v)
	} else {
		end := len(s)
		if end > 0 && s[end-1] == '%' {
			end--
		}
		v, e := strconv.ParseUint(s[:end], 10, 64)
		if e != nil {
			err = fmt.Errorf("percent: %w", e)
			return
		}
		pct = uint(v)
	}
	return
}

// parseGr3d reads one token and parses "X%@[Y,Y,...]" or "X%".
// Returns off=true if the token is "off".
func parseGr3d(sc *scanner) (Gr3d, bool, error) {
	s, ok := sc.next()
	if !ok {
		return Gr3d{}, false, fmt.Errorf("unexpected end")
	}
	if s == "off" {
		return Gr3d{}, true, nil
	}

	at := strings.Index(s, "@[")
	if at < 0 {
		pct_end := len(s)
		if pct_end > 0 && s[pct_end-1] == '%' {
			pct_end--
		}
		v, err := strconv.ParseUint(s[:pct_end], 10, 64)
		if err != nil {
			return Gr3d{}, false, fmt.Errorf("percent: %w", err)
		}
		return Gr3d{Percent: uint(v)}, false, nil
	}

	pct_end := at
	if pct_end > 0 && s[pct_end-1] == '%' {
		pct_end--
	}
	pct, err := strconv.ParseUint(s[:pct_end], 10, 64)
	if err != nil {
		return Gr3d{}, false, fmt.Errorf("percent: %w", err)
	}

	inner, _, ok := strings.Cut(s[at+2:], "]")
	if !ok {
		return Gr3d{}, false, fmt.Errorf("missing ']'")
	}

	n := strings.Count(inner, ",") + 1
	freqs := make([]uint, 0, n)
	for len(inner) > 0 {
		var part string
		part, inner, _ = strings.Cut(inner, ",")
		v, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return Gr3d{}, false, fmt.Errorf("freq: %w", err)
		}
		freqs = append(freqs, uint(v))
	}
	return Gr3d{Percent: uint(pct), Freq: freqs}, false, nil
}

// parsePva reads one token and parses "[X%,...]@Y" or "@Y".
// Returns off=true if the token is "off".
func parsePva(sc *scanner) (Pva, bool, error) {
	s, ok := sc.next()
	if !ok {
		return Pva{}, false, fmt.Errorf("unexpected end")
	}
	if s == "off" {
		return Pva{}, true, nil
	}

	if s[0] == '@' {
		freq, err := strconv.ParseUint(s[1:], 10, 64)
		if err != nil {
			return Pva{}, false, fmt.Errorf("freq: %w", err)
		}
		return Pva{Freq: uint(freq)}, false, nil
	}
	if s[0] != '[' {
		return Pva{}, false, fmt.Errorf("expected '[': %s", s)
	}

	inner, freq_str, ok := strings.Cut(s[1:], "]@")
	if !ok {
		return Pva{}, false, fmt.Errorf("missing ']@': %s", s)
	}

	freq, err := strconv.ParseUint(freq_str, 10, 64)
	if err != nil {
		return Pva{}, false, fmt.Errorf("freq: %w", err)
	}

	n := strings.Count(inner, ",") + 1
	percents := make([]uint, 0, n)
	for len(inner) > 0 {
		var part string
		part, inner, _ = strings.Cut(inner, ",")
		pct_end := len(part)
		if pct_end > 0 && part[pct_end-1] == '%' {
			pct_end--
		}
		v, err := strconv.ParseUint(part[:pct_end], 10, 64)
		if err != nil {
			return Pva{}, false, fmt.Errorf("percent: %w", err)
		}
		percents = append(percents, uint(v))
	}
	return Pva{Percent: percents, Freq: uint(freq)}, false, nil
}
