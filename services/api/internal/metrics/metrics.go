// Package metrics computes the aggregates the dashboard and the coach see (docs/design.md, "Metryki liczone
// w kodzie"). Pure functions over plain values; dates are calendar days in UTC at midnight.
package metrics

import (
	"math"
	"time"
)

const day = 24 * time.Hour

// Days returns every calendar day from..to inclusive.
func Days(from, to time.Time) []time.Time {
	var out []time.Time
	for d := from; !d.After(to); d = d.Add(day) {
		out = append(out, d)
	}
	return out
}

// WeekStart returns the Monday of the ISO week containing d.
func WeekStart(d time.Time) time.Time {
	offset := (int(d.Weekday()) + 6) % 7 // Monday = 0
	return d.Add(-time.Duration(offset) * day)
}

// LoadPoint is the training load state at the end of a day.
type LoadPoint struct {
	Date time.Time
	ATL  float64 // fatigue: exponentially weighted daily load, 7-day time constant
	CTL  float64 // fitness: same, 42-day time constant
}

const (
	ATLDays = 7
	CTLDays = 42
)

// LoadSeries runs the ATL/CTL exponential averages over daily load (Garmin training load per activity, summed
// per day) from `from` to `to`. Days without activities count as zero load. Start `from` well before the
// displayed range: CTL needs a few time constants of history to settle.
func LoadSeries(daily map[time.Time]float64, from, to time.Time) []LoadPoint {
	aATL := 1 - math.Exp(-1.0/ATLDays)
	aCTL := 1 - math.Exp(-1.0/CTLDays)
	var atl, ctl float64
	out := make([]LoadPoint, 0, int(to.Sub(from)/day)+1)
	for _, d := range Days(from, to) {
		load := daily[d]
		atl += (load - atl) * aATL
		ctl += (load - ctl) * aCTL
		out = append(out, LoadPoint{Date: d, ATL: atl, CTL: ctl})
	}
	return out
}

// At returns the point for date d, or false.
func At(series []LoadPoint, d time.Time) (LoadPoint, bool) {
	for i := len(series) - 1; i >= 0; i-- {
		if series[i].Date.Equal(d) {
			return series[i], true
		}
	}
	return LoadPoint{}, false
}

// TSB on day d is form going into that day: yesterday's fitness minus yesterday's fatigue,
// so today's training does not make the morning number worse.
func TSB(series []LoadPoint, d time.Time) (float64, bool) {
	p, ok := At(series, d.Add(-day))
	if !ok {
		return 0, false
	}
	return p.CTL - p.ATL, true
}

// Form buckets TSB for the label next to the number.
type Form string

const (
	Fresh       Form = "fresh"       // >= 0: rested
	Neutral     Form = "neutral"     // -10..0: normal training
	Tired       Form = "tired"       // -30..-10: productive fatigue, watch recovery
	Overreached Form = "overreached" // < -30
)

func FormOf(tsb float64) Form {
	switch {
	case tsb >= 0:
		return Fresh
	case tsb >= -10:
		return Neutral
	case tsb >= -30:
		return Tired
	default:
		return Overreached
	}
}

// Mean of the non-nil values; false when there are fewer than min values (a single reading is noise).
func Mean(values []*float64, min int) (float64, bool) {
	var sum float64
	n := 0
	for _, v := range values {
		if v != nil {
			sum += *v
			n++
		}
	}
	if n == 0 || n < min {
		return 0, false
	}
	return sum / float64(n), true
}

// Round to n decimals.
func Round(v float64, n int) float64 {
	p := math.Pow(10, float64(n))
	return math.Round(v*p) / p
}
