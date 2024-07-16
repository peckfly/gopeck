package biz

import (
	"fmt"
	"math"
	"time"
)

// Pacer refer: https://github.com/tsenart/vegeta/blob/master/lib/pacer.go
// Pacer A Pacer defines the rate of hits during an Attack.
type Pacer interface {
	// Pace returns the duration an Attacker should wait until
	// hitting the next Target, given an already elapsed duration and
	// completed hits. If the second return value is true, an attacker
	// should stop sending hits.
	Pace(elapsed time.Duration, hits uint64) (wait time.Duration, stop bool)

	// Rate returns a Pacer's instantaneous hit rate (per seconds)
	// at the given elapsed duration of an attack.
	Rate(elapsed time.Duration) float64
}

// A PacerFunc is a function adapter type that implements
// the Pacer interface.
type PacerFunc func(time.Duration, uint64) (time.Duration, bool)

// Pace implements the Pacer interface.
func (pf PacerFunc) Pace(elapsed time.Duration, hits uint64) (time.Duration, bool) {
	return pf(elapsed, hits)
}

// A ConstantPacer defines a constant rate of hits for the target.
type ConstantPacer struct {
	Freq int           // Frequency (number of occurrences) per ...
	Per  time.Duration // Time unit, usually 1s
}

// Rate is a type alias for ConstantPacer for backwards-compatibility.
type Rate = ConstantPacer

// ConstantPacer satisfies the Pacer interface.
var _ Pacer = ConstantPacer{}

// String returns a pretty-printed description of the ConstantPacer's behaviour:
//
//	ConstantPacer{Freq: 1, Per: time.Second} => Constant{1 hits/1s}
func (cp ConstantPacer) String() string {
	return fmt.Sprintf("Constant{%d hits/%s}", cp.Freq, cp.Per)
}

// Pace determines the length of time to sleep until the next hit is sent.
func (cp ConstantPacer) Pace(elapsed time.Duration, hits uint64) (time.Duration, bool) {
	switch {
	case cp.Per == 0 || cp.Freq == 0:
		return 0, false // Zero value = infinite rate
	case cp.Per < 0 || cp.Freq < 0:
		return 0, true
	}

	expectedHits := uint64(cp.Freq) * uint64(elapsed/cp.Per)
	if hits < expectedHits {
		// Running behind, send next hit immediately.
		return 0, false
	}
	interval := uint64(cp.Per.Nanoseconds() / int64(cp.Freq))
	if math.MaxInt64/interval < hits {
		// We would overflow delta if we continued, so stop the attack.
		return 0, true
	}
	delta := time.Duration((hits + 1) * interval)
	// Zero or negative durations cause time.Sleep to return immediately.
	return delta - elapsed, false
}

// Rate returns a ConstantPacer's instantaneous hit rate (i.e. requests per second)
// at the given elapsed duration of an attack. Since it's constant, the return
// value is independent of the given elapsed duration.
func (cp ConstantPacer) Rate(elapsed time.Duration) float64 {
	return cp.hitsPerNs() * 1e9
}

// hitsPerNs returns the attack rate this ConstantPacer represents, in
// fractional hits per nanosecond.
func (cp ConstantPacer) hitsPerNs() float64 {
	return float64(cp.Freq) / float64(cp.Per)
}

// LinearPacer paces an attack by starting at a given request rate
// and increasing linearly with the given slope.
type LinearPacer struct {
	StartAt Rate
	Slope   float64
}

// Pace determines the length of time to sleep until the next hit is sent.
func (p LinearPacer) Pace(elapsed time.Duration, hits uint64) (time.Duration, bool) {
	switch {
	case p.StartAt.Per == 0 || p.StartAt.Freq == 0:
		return 0, false // Zero value = infinite rate
	case p.StartAt.Per < 0 || p.StartAt.Freq < 0:
		return 0, true
	}

	expectedHits := p.hits(elapsed)
	if hits == 0 || hits < uint64(expectedHits) {
		// Running behind, send next hit immediately.
		return 0, false
	}

	rate := p.Rate(elapsed)
	interval := math.Round(1e9 / rate)

	if n := uint64(interval); n != 0 && math.MaxInt64/n < hits {
		// We would overflow wait if we continued, so stop the attack.
		return 0, true
	}

	delta := float64(hits+1) - expectedHits
	wait := time.Duration(interval * delta)

	return wait, false
}

// Rate returns a LinearPacer's instantaneous hit rate (i.e. requests per second)
// at the given elapsed duration of an attack.
func (p LinearPacer) Rate(elapsed time.Duration) float64 {
	a := p.Slope
	x := elapsed.Seconds()
	b := p.StartAt.hitsPerNs() * 1e9
	return a*x + b
}

// hits returns the number of hits that have been sent during an attack
// lasting t nanoseconds. It returns a float so we can tell exactly how
// much we've missed our target by when solving numerically in Pace.
func (p LinearPacer) hits(t time.Duration) float64 {
	if t < 0 {
		return 0
	}

	a := p.Slope
	b := p.StartAt.hitsPerNs() * 1e9
	x := t.Seconds()

	return (a*math.Pow(x, 2))/2 + b*x
}
