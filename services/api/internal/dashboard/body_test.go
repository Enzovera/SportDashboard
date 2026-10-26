package dashboard

import (
	"reflect"
	"testing"
	"time"
)

func weighIn(date string, kg float64, muscle *float64) WeighInRow {
	return WeighInRow{MeasuredAt: d(date).Add(5 * time.Hour), LocalDate: d(date), WeightKg: kg, MuscleKg: muscle}
}

// bodyData puts weigh-ins both in WeighIns and as day averages in Days, as the projector does.
func bodyData(ws ...WeighInRow) Data {
	data := Data{WeighIns: ws}
	for _, w := range ws {
		data.Days = append(data.Days, DayRow{Date: w.LocalDate, WeightKg: fp(w.WeightKg)})
	}
	return data
}

func TestBuildBody(t *testing.T) {
	tests := []struct {
		name  string
		data  Data
		check func(t *testing.T, b BodyResponse)
	}{
		{
			name: "no weigh-ins",
			data: Data{},
			check: func(t *testing.T, b BodyResponse) {
				c := b.Composition
				if b.From != "2026-08-03" || b.To != "2026-10-25" || len(b.Weeks) != 12 || c.MeasuredAt != nil || c.Avg7dKg != nil ||
					b.TrendKgPerWeek != nil || len(b.WeighIns) != 0 || c.WeekCount != 0 || c.WeekDays != 7 || len(c.WeekMissing) != 7 {
					t.Errorf("unexpected: %+v", b)
				}
			},
		},
		{
			name: "sparse manual weigh-ins: no average, the few points listed",
			data: bodyData(weighIn("2026-08-21", 78.4, nil), weighIn("2026-09-16", 77.6, nil), weighIn("2026-10-14", 77.0, nil)),
			check: func(t *testing.T, b BodyResponse) {
				c := b.Composition
				if *c.LastKg != 77.0 || c.Avg7dKg != nil || c.RangeCount != 3 || len(b.WeighIns) != 3 || b.WeighIns[0].WeightKg != 77.0 {
					t.Errorf("unexpected: %+v %+v", c, b.WeighIns)
				}
				if b.Weeks[2].Count != 1 || *b.Weeks[2].AvgKg != 78.4 || b.Weeks[0].AvgKg != nil || b.TrendKgPerWeek != nil {
					t.Errorf("weeks: %+v trend %v", b.Weeks, b.TrendKgPerWeek)
				}
			},
		},
		{
			name: "daily weigh-ins: average, change, trend, missing days, list of the last 7 days",
			data: func() Data {
				var ws []WeighInRow
				for i := 83; i >= 0; i-- {
					if i == 4 { // 2026-10-21 skipped
						continue
					}
					date := today.Add(-time.Duration(i) * day)
					kg := 78.0 - 0.02*float64(83-i) // -0.14 kg/week
					muscle := 61.0
					if i < 7 {
						muscle = 61.5
					}
					ws = append(ws, weighIn(iso(date), round1(kg), fp(muscle)))
				}
				return bodyData(ws...)
			}(),
			check: func(t *testing.T, b BodyResponse) {
				c := b.Composition
				if *c.LastKg != 76.3 || *c.Avg7dKg != 76.4 || *c.Avg7dChangeKg != -1.5 || *c.MuscleTrend != "up" || *c.MuscleKg != 61.5 {
					t.Errorf("composition: last %v avg %v change %v muscle %v %v", *c.LastKg, *c.Avg7dKg, *c.Avg7dChangeKg, *c.MuscleKg, *c.MuscleTrend)
				}
				if c.WeekCount != 6 || !reflect.DeepEqual(c.WeekMissing, []string{"2026-10-21"}) {
					t.Errorf("week: %d %v", c.WeekCount, c.WeekMissing)
				}
				if *b.TrendKgPerWeek != -0.14 {
					t.Errorf("trend = %v", *b.TrendKgPerWeek)
				}
				if len(b.WeighIns) != 7 || b.WeighIns[0].MeasuredAt != "2026-10-25T06:00:00+01:00" || b.WeighIns[6].MeasuredAt[:10] != "2026-10-18" {
					t.Errorf("weigh-ins: %+v", b.WeighIns)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { tt.check(t, BuildBody(tt.data, today, now)) })
	}
}
