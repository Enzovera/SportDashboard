package mcpserver

import (
	"strings"
	"testing"
	"time"

	"forma/api/internal/dashboard"
)

func d(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func ip(v int) *int         { return &v }
func fp(v float64) *float64 { return &v }
func sp(v string) *string   { return &v }

var (
	today = d("2026-10-25") // Sunday; week 19–25.10
	now   = time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC)
)

func run(id, date string, km, load float64) dashboard.ActivityRow {
	return dashboard.ActivityRow{ID: id, LocalDate: d(date), StartedAt: d(date).Add(7 * time.Hour), Sport: "running",
		Name: sp("Bieg " + id), DistanceM: fp(km * 1000), DurationS: fp(km * 312), AvgHR: ip(140), Load: fp(load)}
}

func sample() dashboard.Data {
	data := dashboard.Data{
		Activities: []dashboard.ActivityRow{
			run("a1", "2026-10-12", 10, 100),
			run("a2", "2026-10-19", 6, 50),
			{ID: "t1", LocalDate: d("2026-10-20"), Sport: "treadmill_running", DistanceM: fp(5000), DurationS: fp(1500), Load: fp(60)},
			{ID: "c1", LocalDate: d("2026-10-21"), Sport: "cycling", DistanceM: fp(30000), Load: fp(80)},
			run("a4", "2026-10-25", 14, 160),
		},
		Workouts: []dashboard.WorkoutRow{
			{ID: "w1", Title: sp("Push"), LocalDate: d("2026-10-19"), WorkingSets: 18, TonnageKg: 8420.44,
				Muscles: map[string]int{"chest": 9, "shoulders": 6, "triceps": 3}},
			{ID: "w2", Title: sp("Legs"), LocalDate: d("2026-10-21"), WorkingSets: 16, TonnageKg: 11950,
				Muscles: map[string]int{"quadriceps": 7, "hamstrings": 5, "chest": 4}},
			{ID: "w0", Title: sp("Pull"), LocalDate: d("2026-10-16"), WorkingSets: 12, TonnageKg: 5000,
				Muscles: map[string]int{"upper_back": 12}},
		},
	}
	for i := 59; i >= 0; i-- {
		dt := today.Add(-time.Duration(i) * day)
		row := dashboard.DayRow{Date: dt, HRVMs: fp(54), SleepS: ip(27000), RestingHR: ip(49)}
		if i < 7 {
			row.HRVMs, row.WeightKg = fp(52), fp(78+float64(i)*0.1)
		}
		data.Days = append(data.Days, row)
	}
	return data
}

func TestParseRange(t *testing.T) {
	cases := []struct {
		name, from, to   string
		wantFrom, wantTo string
		err              string
	}{
		{name: "defaults", wantFrom: "2026-09-28", wantTo: "2026-10-25"},
		{name: "explicit", from: "2026-10-01", to: "2026-10-07", wantFrom: "2026-10-01", wantTo: "2026-10-07"},
		{name: "from only", from: "2026-10-20", wantFrom: "2026-10-20", wantTo: "2026-10-25"},
		{name: "reversed", from: "2026-10-08", to: "2026-10-07", err: "after"},
		{name: "too long", from: "2025-01-01", to: "2026-10-25", err: "exceeds"},
		{name: "bad format", from: "25.10.2026", err: "YYYY-MM-DD"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, to, err := parseRange(c.from, c.to, today)
			if c.err != "" {
				if err == nil || !strings.Contains(err.Error(), c.err) {
					t.Fatalf("err = %v, want %q", err, c.err)
				}
				return
			}
			if err != nil || iso(f) != c.wantFrom || iso(to) != c.wantTo {
				t.Errorf("got %s..%s, %v; want %s..%s", iso(f), iso(to), err, c.wantFrom, c.wantTo)
			}
		})
	}
}

func TestBuildDaySummary(t *testing.T) {
	detail := &DayDetail{Date: today, SleepS: ip(26100), SleepScore: ip(81), HRVMs: ip(51), RestingHR: ip(48), Steps: ip(9000), WeightKg: fp(78.04)}
	got := BuildDaySummary(sample(), detail, today, now)

	if got.SleepH == nil || *got.SleepH != 7.25 || *got.WeightKg != 78 || *got.HRV7dMs != 52 {
		t.Errorf("sleep/weight/hrv: %+v", got)
	}
	// 60-day baseline: 53 nights at 54 ms, 7 at 52 ms.
	if got.HRVBaselineMs == nil || *got.HRVBaselineMs != 53.8 {
		t.Errorf("HRVBaselineMs = %v", got.HRVBaselineMs)
	}
	if *got.Weight7dKg != 78.3 {
		t.Errorf("Weight7dKg = %v", *got.Weight7dKg)
	}
	if len(got.Activities) != 1 || *got.Activities[0].PaceMinPerKm != "5:12" || *got.Activities[0].DistanceKm != 14 {
		t.Errorf("activities: %+v", got.Activities)
	}
	if got.Load.TSB == nil || got.Load.Form == nil || got.Load.ATL == nil {
		t.Errorf("load: %+v", got.Load)
	}
	if len(got.Sync) != 2 || len(got.Missing) != 1 {
		t.Errorf("sync %v, missing %v", got.Sync, got.Missing)
	}

	empty := BuildDaySummary(dashboard.Data{}, nil, today, now)
	if empty.SleepH != nil || len(empty.Missing) != 2 || empty.Activities == nil || empty.Workouts == nil {
		t.Errorf("no data: %+v", empty)
	}
}

func TestBuildActivities(t *testing.T) {
	cases := []struct {
		sport string
		want  []string // sports, in order
	}{
		{"", []string{"running", "treadmill_running", "cycling", "running"}},
		{"running", []string{"running", "treadmill_running", "running"}},
		{"cycling", []string{"cycling"}},
	}
	for _, c := range cases {
		got := BuildActivities(sample(), d("2026-10-19"), today, c.sport)
		var sports []string
		for _, a := range got.Activities {
			sports = append(sports, a.Sport)
		}
		if strings.Join(sports, ",") != strings.Join(c.want, ",") {
			t.Errorf("sport %q: got %v, want %v", c.sport, sports, c.want)
		}
	}
	tm := BuildActivities(sample(), d("2026-10-20"), d("2026-10-20"), "")
	if a := tm.Activities[0]; *a.PaceMinPerKm != "5:00" || *a.DurationMin != 25 {
		t.Errorf("treadmill: %+v", a)
	}
	if c := BuildActivities(sample(), d("2026-10-21"), d("2026-10-21"), "").Activities[0]; c.PaceMinPerKm != nil {
		t.Errorf("cycling has pace %v", *c.PaceMinPerKm)
	}
}

func TestBuildStrengthVolume(t *testing.T) {
	week := BuildStrengthVolume(sample(), d("2026-10-14"), today, "week")
	if len(week.Rows) != 2 {
		t.Fatalf("weeks: %+v", week.Rows)
	}
	// From Wednesday 14.10: whole ISO weeks, starting on Monday 12.10.
	if r := week.Rows[0]; r.Key != "2026-10-12" || r.Sessions != 1 || r.WorkingSets != 12 || *r.TonnageKg != 5000 {
		t.Errorf("week 12.10: %+v", r)
	}
	if r := week.Rows[1]; r.Sessions != 2 || r.WorkingSets != 34 || *r.TonnageKg != 20370.4 || r.SetsByMuscle["chest"] != 13 {
		t.Errorf("week 19.10: %+v", r)
	}

	muscle := BuildStrengthVolume(sample(), d("2026-10-19"), today, "muscle")
	if r := muscle.Rows[0]; r.Key != "chest" || r.WorkingSets != 13 || r.Sessions != 2 || r.TonnageKg != nil {
		t.Errorf("muscle top: %+v", r)
	}

	session := BuildStrengthVolume(sample(), d("2026-10-19"), today, "session")
	if len(session.Rows) != 2 || *session.Rows[0].Title != "Push" || *session.Rows[0].TonnageKg != 8420.4 {
		t.Errorf("sessions: %+v", session.Rows)
	}
}

func TestBuildTrend(t *testing.T) {
	cases := []struct {
		metric, from string
		gran         string
		want         []*float64
	}{
		{"run_km_week", "2026-10-14", "week", []*float64{fp(10), fp(25)}},
		{"long_run_km_week", "2026-10-19", "week", []*float64{fp(14)}},
		{"strength_sets_week", "2026-10-12", "week", []*float64{fp(12), fp(34)}},
		{"training_load", "2026-10-19", "day", []*float64{fp(50), fp(60), fp(80), fp(0), fp(0), fp(0), fp(160)}},
		{"weight", "2026-10-17", "day", []*float64{nil, nil, fp(78.6), fp(78.5), fp(78.4), fp(78.3), fp(78.2), fp(78.1), fp(78)}},
		{"weight_7d", "2026-10-25", "day", []*float64{fp(78.3)}},
		{"hrv_7d", "2026-10-25", "day", []*float64{fp(52)}},
		{"sleep_h", "2026-10-25", "day", []*float64{fp(7.5)}},
	}
	for _, c := range cases {
		t.Run(c.metric, func(t *testing.T) {
			got, err := BuildTrend(sample(), c.metric, d(c.from), today)
			if err != nil {
				t.Fatal(err)
			}
			if got.Granularity != c.gran || len(got.Points) != len(c.want) {
				t.Fatalf("got %s %+v", got.Granularity, got.Points)
			}
			for i, p := range got.Points {
				if (p.Value == nil) != (c.want[i] == nil) || (p.Value != nil && *p.Value != *c.want[i]) {
					t.Errorf("point %s = %v, want %v", p.Date, deref(p.Value), deref(c.want[i]))
				}
			}
		})
	}
	daily, _ := BuildTrend(sample(), "hrv", today, today)
	if daily.Points[0].ChangePct != nil {
		t.Error("daily metric has change_pct")
	}
	if _, err := BuildTrend(sample(), "vo2max", today, today); err == nil {
		t.Error("unknown metric accepted")
	}
}

func TestWeeklyChangePct(t *testing.T) {
	cases := []struct {
		metric, from string
		want         []*float64 // change_pct per week
	}{
		// 5.10: 0 km, 12.10: 10 km, 19.10: 25 km.
		{"run_km_week", "2026-10-05", []*float64{nil, nil, fp(150)}},
		// 12.10: 12 sets, 19.10: 34 sets.
		{"strength_sets_week", "2026-10-12", []*float64{nil, fp(183)}},
		// 19.10 vs 12.10: longest run 14 km vs 10 km.
		{"long_run_km_week", "2026-10-19", []*float64{fp(40)}},
	}
	for _, c := range cases {
		t.Run(c.metric, func(t *testing.T) {
			got, _ := BuildTrend(sample(), c.metric, d(c.from), today)
			if len(got.Points) != len(c.want) {
				t.Fatalf("points: %+v", got.Points)
			}
			for i, p := range got.Points {
				if (p.ChangePct == nil) != (c.want[i] == nil) || (p.ChangePct != nil && *p.ChangePct != *c.want[i]) {
					t.Errorf("%s change_pct = %v, want %v", p.Date, deref(p.ChangePct), deref(c.want[i]))
				}
			}
		})
	}
	for _, c := range []struct {
		cur, prev float64
		want      *float64
	}{
		{1, 9.6, fp(-90)},
		{0, 1, fp(-100)},
		{5, 0, nil},
	} {
		if got := changePct(c.cur, true, c.prev, true); deref(got) != deref(c.want) {
			t.Errorf("changePct(%v, %v) = %v, want %v", c.cur, c.prev, deref(got), deref(c.want))
		}
	}
}

func deref(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

func TestTSBIsMorningForm(t *testing.T) {
	data := sample()
	tsb, _ := BuildTrend(data, "tsb", today, today)
	atl, _ := BuildTrend(data, "atl", today.Add(-day), today.Add(-day))
	ctl, _ := BuildTrend(data, "ctl", today.Add(-day), today.Add(-day))
	// Rounding each side separately may differ by 0.1.
	if diff := *tsb.Points[0].Value - (*ctl.Points[0].Value - *atl.Points[0].Value); diff > 0.11 || diff < -0.11 {
		t.Errorf("tsb %v != ctl %v - atl %v of the day before", *tsb.Points[0].Value, *ctl.Points[0].Value, *atl.Points[0].Value)
	}
}

func TestValidateNote(t *testing.T) {
	cases := []struct {
		name, kind, date, content string
		wantDate, err             string
	}{
		{name: "default date", kind: "daily", content: "  Dobry dzień.  ", wantDate: "2026-10-25"},
		{name: "next week", kind: "weekly", date: "2026-11-01", content: "Plan", wantDate: "2026-11-01"},
		{name: "unknown kind", kind: "plan", content: "x", err: "kind"},
		{name: "empty", kind: "decision", content: "   ", err: "content"},
		{name: "too long", kind: "decision", content: strings.Repeat("ą", maxNoteChars+1), err: "content"},
		{name: "too old", kind: "daily", date: "2026-09-01", content: "x", err: "within"},
		{name: "too far ahead", kind: "daily", date: "2026-11-10", content: "x", err: "within"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			date, content, err := validateNote(c.kind, c.date, c.content, today)
			if c.err != "" {
				if err == nil || !strings.Contains(err.Error(), c.err) {
					t.Fatalf("err = %v, want %q", err, c.err)
				}
				return
			}
			if err != nil || iso(date) != c.wantDate || content != strings.TrimSpace(c.content) {
				t.Errorf("got %s %q %v", iso(date), content, err)
			}
		})
	}
}
