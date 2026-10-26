package dashboard

import (
	"time"

	"forma/api/internal/metrics"
	"forma/api/internal/projector"
)

const (
	bodyRangeWeeks = 12
	// Weigh-ins listed: everything from the last 7 days, but at least the last 7 weigh-ins.
	bodyListDays = 7
	bodyListMin  = 7
	// Weekly trend needs this many weeks with weigh-ins in the range.
	trendMinWeeks = 4
	// Muscle mass change (kg) between the last 7 days and 4 weeks earlier that still reads as stable.
	muscleStableKg = 0.3
)

// ---------- /api/body ----------

func BuildBody(data Data, today, now time.Time) BodyResponse {
	thisWeek := metrics.WeekStart(today)
	from := thisWeek.Add(-(bodyRangeWeeks - 1) * 7 * day)
	resp := BodyResponse{
		Date:       iso(today),
		From:       iso(from),
		To:         iso(thisWeek.Add(6 * day)),
		RangeWeeks: bodyRangeWeeks,
		Sync:       syncOf(BuildSync(data.Sync, now), "garmin"),
		WeighIns:   []WeighIn{},
	}

	var inRange []WeighInRow
	for _, w := range data.WeighIns {
		if !w.LocalDate.Before(from) && !w.LocalDate.After(today) {
			inRange = append(inRange, w)
		}
	}
	resp.Composition = buildComposition(data, inRange, today, from)
	resp.Weeks = bodyWeeks(inRange, from, thisWeek)
	resp.TrendKgPerWeek = weeklyTrend(resp.Weeks)

	// Newest first: the last 7 days, or the last 7 weigh-ins when that reaches further back.
	for i := len(data.WeighIns) - 1; i >= 0; i-- {
		w := data.WeighIns[i]
		if w.LocalDate.After(today) {
			continue
		}
		if len(resp.WeighIns) >= bodyListMin && w.LocalDate.Before(today.Add(-(bodyListDays-1)*day)) {
			break
		}
		resp.WeighIns = append(resp.WeighIns, WeighIn{
			MeasuredAt: w.MeasuredAt.In(projector.Warsaw).Format(time.RFC3339),
			WeightKg:   w.WeightKg, FatPct: w.FatPct, MuscleKg: w.MuscleKg, WaterPct: w.WaterPct,
		})
	}
	return resp
}

// syncOf keeps the sources a view depends on.
func syncOf(all []SyncStatus, sources ...string) []SyncStatus {
	out := []SyncStatus{}
	for _, s := range all {
		if contains(sources, s.Source) {
			out = append(out, s)
		}
	}
	return out
}

func buildComposition(data Data, inRange []WeighInRow, today, from time.Time) BodyComposition {
	c := BodyComposition{RangeCount: len(inRange), WeekMissing: []string{}}

	// Latest values: the last weigh-in, and the last composition reading (the scale does not always send it).
	for i := len(data.WeighIns) - 1; i >= 0; i-- {
		w := data.WeighIns[i]
		if w.LocalDate.After(today) {
			continue
		}
		if c.MeasuredAt == nil {
			c.MeasuredAt = ptr(w.MeasuredAt.In(projector.Warsaw).Format(time.RFC3339))
			c.LastKg = ptr(w.WeightKg)
		}
		if c.FatPct == nil && w.FatPct != nil {
			c.FatPct = w.FatPct
		}
		if c.MuscleKg == nil && w.MuscleKg != nil {
			c.MuscleKg = ptr(round1(*w.MuscleKg))
		}
	}

	// 7-day mean, the same as on the Today view (day averages, at least 3 weigh-ins).
	if t := buildWeight(data.Days, today); t != nil {
		c.Avg7dKg = ptr(t.Avg7dKg)
		// Change across the range: against the first 7-day mean in it that has enough weigh-ins.
		for d := from.Add(6 * day); d.Before(today); d = d.Add(day) {
			if first := buildWeight(data.Days, d); first != nil {
				c.Avg7dChangeKg = ptr(round1(t.Avg7dKg - first.Avg7dKg))
				break
			}
		}
	}

	c.MuscleTrend = muscleTrend(data.WeighIns, today)

	weekStart := metrics.WeekStart(today)
	weighed := map[time.Time]bool{}
	for _, w := range inRange {
		weighed[w.LocalDate] = true
	}
	for _, d := range metrics.Days(weekStart, today) {
		c.WeekDays++
		if weighed[d] {
			c.WeekCount++
		} else {
			c.WeekMissing = append(c.WeekMissing, iso(d))
		}
	}
	return c
}

// muscleTrend compares the mean muscle mass of the last 7 days with the same window 4 weeks earlier.
func muscleTrend(ws []WeighInRow, today time.Time) *string {
	muscle := func(from, to time.Time) (float64, bool) {
		var vals []*float64
		for _, w := range ws {
			if !w.LocalDate.Before(from) && !w.LocalDate.After(to) {
				vals = append(vals, w.MuscleKg)
			}
		}
		return metrics.Mean(vals, 1)
	}
	now, ok1 := muscle(today.Add(-6*day), today)
	before, ok2 := muscle(today.Add(-34*day), today.Add(-28*day))
	if !ok1 || !ok2 {
		return nil
	}
	switch diff := now - before; {
	case diff >= muscleStableKg:
		return ptr("up")
	case diff <= -muscleStableKg:
		return ptr("down")
	default:
		return ptr("stable")
	}
}

func bodyWeeks(ws []WeighInRow, from, thisWeek time.Time) []BodyWeek {
	var out []BodyWeek
	for start := from; !start.After(thisWeek); start = start.Add(7 * day) {
		bw := BodyWeek{WeekStart: iso(start)}
		var sum float64
		for _, w := range ws {
			if w.LocalDate.Before(start) || !w.LocalDate.Before(start.Add(7*day)) {
				continue
			}
			bw.Count++
			sum += w.WeightKg
			if bw.MinKg == nil || w.WeightKg < *bw.MinKg {
				bw.MinKg = ptr(w.WeightKg)
			}
			if bw.MaxKg == nil || w.WeightKg > *bw.MaxKg {
				bw.MaxKg = ptr(w.WeightKg)
			}
		}
		if bw.Count > 0 {
			bw.AvgKg = ptr(round1(sum / float64(bw.Count)))
			bw.MinKg, bw.MaxKg = ptr(round1(*bw.MinKg)), ptr(round1(*bw.MaxKg))
		}
		out = append(out, bw)
	}
	return out
}

// weeklyTrend is the least-squares slope of the weekly averages, kg per week; null with fewer than 4 weeks of data.
func weeklyTrend(weeks []BodyWeek) *float64 {
	var xs, ys []float64
	for i, w := range weeks {
		if w.AvgKg != nil {
			xs = append(xs, float64(i))
			ys = append(ys, *w.AvgKg)
		}
	}
	slope, ok := metrics.Slope(xs, ys, trendMinWeeks)
	if !ok {
		return nil
	}
	return ptr(metrics.Round(slope, 2))
}
