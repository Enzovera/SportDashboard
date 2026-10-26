package dashboard

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func d(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func ip(v int) *int             { return &v }
func fp(v float64) *float64     { return &v }
func sp(v string) *string       { return &v }
func tp(v time.Time) *time.Time { return &v }

var (
	today = d("2026-10-25") // Sunday; week 19–25.10
	now   = time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC)
)

func run(id, date string, km, load float64) ActivityRow {
	return ActivityRow{ID: id, LocalDate: d(date), StartedAt: d(date).Add(7 * time.Hour), Sport: "running",
		Name: sp("Bieg " + id), DistanceM: fp(km * 1000), DurationS: fp(km * 360), AvgHR: ip(140), Load: fp(load)}
}

func sample() Data {
	data := Data{
		Activities: []ActivityRow{
			run("a1", "2026-10-12", 10, 100),
			run("a2", "2026-10-19", 6, 50),
			run("a3", "2026-10-20", 10.2, 130),
			{ID: "c1", LocalDate: d("2026-10-21"), Sport: "cycling", DistanceM: fp(30000), Load: fp(80)},
			run("a4", "2026-10-24", 8, 118),
		},
		Workouts: []WorkoutRow{
			{ID: "w1", Title: sp("Push"), LocalDate: d("2026-10-19"), StartedAt: now, EndedAt: tp(now.Add(58 * time.Minute)),
				WorkingSets: 18, TonnageKg: 8420.4, Muscles: map[string]int{"chest": 9, "shoulders": 6, "triceps": 3}},
			{ID: "w2", Title: sp("Legs"), LocalDate: d("2026-10-21"), StartedAt: now,
				WorkingSets: 16, TonnageKg: 11950, Muscles: map[string]int{"quadriceps": 7, "hamstrings": 5, "glutes": 4, "cardio": 2}},
			{ID: "w0", Title: sp("Pull"), LocalDate: d("2026-10-16"), StartedAt: now, WorkingSets: 17, Muscles: map[string]int{"upper_back": 12}},
		},
		Sync: []SyncRow{
			{Source: "garmin", LastStatus: "ok", LastSuccessAt: tp(now.Add(-time.Hour))},
			{Source: "hevy", LastStatus: "error", LastSuccessAt: tp(now.Add(-15 * time.Hour))},
		},
	}
	for i := 0; i < 60; i++ {
		dt := today.Add(-time.Duration(i) * day)
		row := DayRow{Date: dt, HRVMs: fp(54), SleepS: ip(25800), RestingHR: ip(49)}
		if i < 7 {
			row.HRVMs = fp(52)
		}
		data.Days = append([]DayRow{row}, data.Days...)
	}
	data.Days[len(data.Days)-1].SleepS = ip(27600)
	data.Days[len(data.Days)-1].RestingHR = ip(48)
	return data
}

func TestBuildTodayWeek(t *testing.T) {
	w := BuildToday(sample(), today, now).Week
	if w.From != "2026-10-19" || w.To != "2026-10-25" || len(w.Days) != 7 {
		t.Fatalf("week bounds %s..%s, %d days", w.From, w.To, len(w.Days))
	}
	// Runs only: the bike ride on Wednesday does not count as running mileage.
	if w.RunKmDone != 24.2 || w.StrengthDone != 2 {
		t.Errorf("RunKmDone=%v StrengthDone=%d", w.RunKmDone, w.StrengthDone)
	}
	tests := []struct {
		i        int
		km       *float64
		strength *string
	}{{0, fp(6), sp("Push")}, {1, fp(10.2), nil}, {2, nil, sp("Legs")}, {5, fp(8), nil}, {6, nil, nil}}
	for _, tt := range tests {
		got := w.Days[tt.i]
		if (got.RunKm == nil) != (tt.km == nil) || (got.RunKm != nil && *got.RunKm != *tt.km) ||
			(got.Strength == nil) != (tt.strength == nil) || (got.Strength != nil && *got.Strength != *tt.strength) {
			t.Errorf("day %d = %+v", tt.i, got)
		}
	}
}

func TestBuildTodayReadiness(t *testing.T) {
	r := BuildToday(sample(), today, now).Readiness
	if *r.HRV7dMs != 52 || *r.HRVBaselineMs != 54 || r.HRVTone != "good" || *r.HRVLabel != "w normie" {
		t.Errorf("hrv: %v %v %s", *r.HRV7dMs, *r.HRVBaselineMs, r.HRVTone)
	}
	if *r.SleepS != 27600 || *r.RestingHR != 48 || *r.RestingHRDelta7d != -1 {
		t.Errorf("sleep/rhr: %+v", r)
	}
	if r.TSB == nil || r.Form == nil {
		t.Error("tsb/form missing")
	}

	low := sample()
	for i := len(low.Days) - 7; i < len(low.Days); i++ {
		low.Days[i].HRVMs = fp(45)
	}
	if r := BuildToday(low, today, now).Readiness; r.HRVTone != "warn" || *r.HRVLabel != "poniżej bazy" {
		t.Errorf("low hrv: %s", r.HRVTone)
	}
}

func TestBuildWeight(t *testing.T) {
	tests := []struct {
		name    string
		weights map[string]float64
		want    *WeightTrend
	}{
		{"too few weigh-ins", map[string]float64{"2026-10-25": 76.4, "2026-10-22": 76.6}, nil},
		{"mean without previous week", map[string]float64{"2026-10-25": 76.2, "2026-10-23": 76.4, "2026-10-21": 76.6}, &WeightTrend{Avg7dKg: 76.4}},
		{"with trend", map[string]float64{"2026-10-25": 76.2, "2026-10-23": 76.4, "2026-10-21": 76.6,
			"2026-10-17": 76.7, "2026-10-15": 76.7, "2026-10-13": 76.7}, &WeightTrend{Avg7dKg: 76.4, DeltaPerWeekKg: fp(-0.3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var days []DayRow
			for ds, w := range tt.weights {
				days = append(days, DayRow{Date: d(ds), WeightKg: fp(w)})
			}
			got := buildWeight(days, today)
			if (got == nil) != (tt.want == nil) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			if got == nil {
				return
			}
			if got.Avg7dKg != tt.want.Avg7dKg || (got.DeltaPerWeekKg == nil) != (tt.want.DeltaPerWeekKg == nil) ||
				(got.DeltaPerWeekKg != nil && *got.DeltaPerWeekKg != *tt.want.DeltaPerWeekKg) {
				t.Errorf("got %+v (delta %v)", got, got.DeltaPerWeekKg)
			}
		})
	}
}

func TestBuildSync(t *testing.T) {
	rows := []SyncRow{
		{Source: "garmin", LastStatus: "ok", LastSuccessAt: tp(now.Add(-time.Hour))},
		{Source: "hevy", LastStatus: "ok", LastSuccessAt: tp(now.Add(-4 * time.Hour))},
		{Source: "fitatu", LastStatus: "partial", LastSuccessAt: tp(now.Add(-time.Hour))},
	}
	got := BuildSync(rows, now)
	want := map[string]string{"garmin": "ok", "hevy": "stale", "fitatu": "partial"}
	if len(got) != 3 {
		t.Fatalf("got %+v", got)
	}
	for _, s := range got {
		if s.Status != want[s.Source] {
			t.Errorf("%s = %s, want %s", s.Source, s.Status, want[s.Source])
		}
	}
	if got := BuildSync(nil, now); len(got) != 2 || got[0].Status != "stale" || got[0].LastSuccessAt != nil {
		t.Errorf("no runs: %+v", got)
	}
}

func TestBuildTraining(t *testing.T) {
	tr := BuildTraining(sample(), today, "4w", now)
	if tr.Range != "4w" || tr.From != "2026-09-28" || tr.To != "2026-10-25" || len(tr.WeeklyKm) != 4 {
		t.Fatalf("range %s %s..%s weeks=%d", tr.Range, tr.From, tr.To, len(tr.WeeklyKm))
	}
	if tr.WeeklyKm[2].Km != 10 || tr.WeeklyKm[3].Km != 24.2 {
		t.Errorf("weekly km %+v", tr.WeeklyKm)
	}
	if tr.Load.WeekKm != 24.2 || *tr.Load.PrevWeekKm != 10 || tr.Load.CTL == nil || tr.Load.ATL == nil {
		t.Errorf("load %+v", tr.Load)
	}
	if len(tr.LoadSeries) != 28 || tr.LoadSeries[0].Date != "2026-09-28" {
		t.Errorf("load series len=%d first=%s", len(tr.LoadSeries), tr.LoadSeries[0].Date)
	}
	if len(tr.RecentRuns) != 4 || tr.RecentRuns[0].ID != "a4" || *tr.RecentRuns[0].PaceSPerKm != 360 {
		t.Errorf("runs %+v", tr.RecentRuns)
	}
	// This week's sets per muscle, most first, cardio excluded; last week's Pull not counted.
	items := tr.SetsPerMuscle.Items
	if len(items) != 6 || items[0].Muscle != "chest" || items[0].Sets != 9 {
		t.Errorf("sets %+v", items)
	}
	if len(tr.StrengthSessions) != 2 || tr.StrengthSessions[0].DurationMin != 58 || tr.StrengthSessions[0].TonnageKg != 8420 {
		t.Errorf("strength %+v", tr.StrengthSessions)
	}
	if got := BuildTraining(sample(), today, "nonsense", now); got.Range != "12w" || len(got.WeeklyKm) != 12 {
		t.Errorf("default range %s %d", got.Range, len(got.WeeklyKm))
	}
}

// Later-stage sections must be present as null, never omitted: the UI distinguishes "no data" from "missing key".
func TestTodayJSONHasNullSections(t *testing.T) {
	b, err := json.Marshal(BuildToday(Data{}, today, now))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"coach":null`, `"session":null`, `"tomorrow":null`, `"nutrition":null`,
		`"supplements":null`, `"proposalsPending":null`, `"raceWeeksLeft":null`, `"weight":null`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("missing %s in %s", key, b)
		}
	}
}
