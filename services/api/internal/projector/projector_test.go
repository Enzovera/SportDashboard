package projector

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func dp(s string) *time.Time { t := date(s); return &t }

func TestApplyDayEndpoints(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*Day, []byte) error
		body  string
		check func(Day) bool
	}{
		{"sleep", ApplySleep, `{"dailySleepDTO":{"sleepTimeSeconds":27600,"sleepScores":{"overall":{"value":82}}}}`,
			func(d Day) bool { return *d.SleepS == 27600 && *d.SleepScore == 82 }},
		{"sleep not recorded", ApplySleep, `{"dailySleepDTO":{"sleepTimeSeconds":null}}`,
			func(d Day) bool { return d.SleepS == nil && d.SleepScore == nil }},
		{"hrv", ApplyHRV, `{"hrvSummary":{"lastNightAvg":62,"weeklyAvg":60,"status":"UNBALANCED"}}`,
			func(d Day) bool {
				return *d.HRVLastNightMs == 62 && *d.HRVWeeklyAvgMs == 60 && *d.HRVStatus == "UNBALANCED"
			}},
		{"hrv missing summary", ApplyHRV, `{}`, func(d Day) bool { return d.HRVLastNightMs == nil }},
		{"user summary", ApplyUserSummary, `{"totalSteps":7881,"restingHeartRate":49,"activeKilocalories":615.4,"totalKilocalories":2274.0}`,
			func(d Day) bool {
				return *d.Steps == 7881 && *d.RestingHR == 49 && *d.ActiveKcal == 615 && *d.TotalKcal == 2274
			}},
		{"weight in grams", ApplyBodyComposition, `{"totalAverage":{"weight":76400.0}}`,
			func(d Day) bool { return *d.WeightKg == 76.4 }},
		{"no weigh-in", ApplyBodyComposition, `{"totalAverage":{"weight":null}}`, func(d Day) bool { return d.WeightKg == nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Day
			if err := tt.apply(&d, []byte(tt.body)); err != nil {
				t.Fatal(err)
			}
			if !tt.check(d) {
				t.Errorf("unexpected day: %+v", d)
			}
		})
	}
}

const run = `{"activityId":24711716527,"activityName":"Bytom - Beztlenowe","activityType":{"typeKey":"running"},
 "startTimeLocal":"2026-10-13 00:52:58","startTimeGMT":"2026-10-12 22:52:58","distance":9603.75,"duration":3150.2,
 "movingDuration":3142.4,"averageHR":138.0,"maxHR":161.0,"elevationGain":48.0,"activityTrainingLoad":141.03}`

func TestParseActivity(t *testing.T) {
	a, err := ParseActivity([]byte(run))
	if err != nil {
		t.Fatal(err)
	}
	// Local date comes from startTimeLocal: just after midnight in Warsaw is still the 13th, not the 12th (GMT).
	if a.ExternalID != "24711716527" || !a.LocalDate.Equal(date("2026-10-13")) || !a.IsRun() || *a.AvgHR != 138 {
		t.Errorf("unexpected activity: %+v", a)
	}
	if want := time.Date(2026, 10, 12, 22, 52, 58, 0, time.UTC); !a.StartedAt.Equal(want) {
		t.Errorf("StartedAt = %v, want %v", a.StartedAt, want)
	}
}

func TestIsRun(t *testing.T) {
	for sport, want := range map[string]bool{"running": true, "treadmill_running": true, "trail_running": true,
		"cycling": false, "strength_training": false, "hiking": false} {
		if got := (Activity{Sport: sport}).IsRun(); got != want {
			t.Errorf("IsRun(%s) = %v", sport, got)
		}
	}
}

const workout = `{"id":"w1","title":"Push","start_time":"2026-10-25T23:30:00+00:00","end_time":"2026-10-26T00:30:00+00:00",
 "exercises":[
  {"index":0,"title":"Bench Press","exercise_template_id":"T1","sets":[
    {"index":0,"type":"warmup","weight_kg":40,"reps":10},
    {"index":1,"type":"normal","weight_kg":80,"reps":8,"rpe":8},
    {"index":2,"type":"failure","weight_kg":80,"reps":6}]},
  {"index":1,"title":"Rozgrzewka","exercise_template_id":"X","sets":[{"index":0,"type":"normal","duration_seconds":399}]}]}`

func TestParseWorkout(t *testing.T) {
	chest := "chest"
	w, err := ParseWorkout([]byte(workout), map[string]*string{"T1": &chest})
	if err != nil {
		t.Fatal(err)
	}
	// 23:30 UTC on Oct 25 is already Oct 26 in Warsaw (CET, +1, after the DST change that day).
	if !w.LocalDate.Equal(date("2026-10-26")) {
		t.Errorf("LocalDate = %v", w.LocalDate)
	}
	// Volume counts working sets only: 80*8 + 80*6; the warm-up and the timed set do not count.
	if w.VolumeKg != 1120 {
		t.Errorf("VolumeKg = %v, want 1120", w.VolumeKg)
	}
	working := 0
	for _, s := range w.Sets {
		if s.Working() {
			working++
			if s.MuscleGroup == nil || *s.MuscleGroup != "chest" {
				t.Errorf("set %+v: muscle group not resolved", s)
			}
		}
	}
	if len(w.Sets) != 4 || working != 2 {
		t.Errorf("sets = %d, working = %d", len(w.Sets), working)
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name      string
		in        []payload
		days      int
		acts      int
		workouts  int
		skipped   int
		checkDays func([]Day) bool
	}{
		{
			name: "merges endpoints of one day",
			in: []payload{
				{id: 1, source: "garmin", endpoint: "sleep", externalID: "2026-10-25", forDate: dp("2026-10-25"), body: []byte(`{"dailySleepDTO":{"sleepTimeSeconds":27600}}`)},
				{id: 2, source: "garmin", endpoint: "hrv", externalID: "2026-10-25", forDate: dp("2026-10-25"), body: []byte(`{"hrvSummary":{"lastNightAvg":52}}`)},
				{id: 3, source: "garmin", endpoint: "user_summary", externalID: "2026-10-24", forDate: dp("2026-10-24"), body: []byte(`{"restingHeartRate":48}`)},
			},
			days: 2,
			checkDays: func(d []Day) bool {
				return d[1].Date.Equal(date("2026-10-25")) && *d[1].SleepS == 27600 && *d[1].HRVLastNightMs == 52 && *d[0].RestingHR == 48
			},
		},
		{
			name: "attaches HR zones to activities",
			in: []payload{
				{id: 1, source: "garmin", endpoint: "activities", externalID: "24711716527", body: []byte(run)},
				{id: 2, source: "garmin", endpoint: "activity_hr_zones", externalID: "24711716527", body: []byte(`[{"zoneNumber":2,"secsInZone":600.4}]`)},
			},
			acts: 1,
		},
		{
			name: "drops workouts deleted after their last update",
			in: []payload{
				{id: 1, source: "hevy", endpoint: "workouts", externalID: "w1", body: []byte(workout)},
				{id: 2, source: "hevy", endpoint: "workout_deleted", externalID: "w1", body: []byte(`{"type":"deleted","id":"w1"}`)},
			},
		},
		{
			name: "keeps a workout re-created after a delete",
			in: []payload{
				{id: 1, source: "hevy", endpoint: "workout_deleted", externalID: "w1", body: []byte(`{"type":"deleted","id":"w1"}`)},
				{id: 2, source: "hevy", endpoint: "workouts", externalID: "w1", body: []byte(workout)},
			},
			workouts: 1,
		},
		{
			name: "skips a broken payload without failing the rest",
			in: []payload{
				{id: 1, source: "garmin", endpoint: "activities", externalID: "x", body: []byte(`{"activityId":1,"startTimeGMT":"nope"}`)},
				{id: 2, source: "garmin", endpoint: "activities", externalID: "24711716527", body: []byte(run)},
			},
			acts:    1,
			skipped: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := Build(tt.in)
			if len(p.Days) != tt.days || len(p.Activities) != tt.acts || len(p.Workouts) != tt.workouts || p.Skipped != tt.skipped {
				t.Fatalf("got days=%d acts=%d workouts=%d skipped=%d", len(p.Days), len(p.Activities), len(p.Workouts), p.Skipped)
			}
			if tt.checkDays != nil && !tt.checkDays(p.Days) {
				t.Errorf("unexpected days: %+v", p.Days)
			}
			if tt.acts == 1 && tt.name == "attaches HR zones to activities" && p.Activities[0].HRZonesS[1] != 600 {
				t.Errorf("zones = %v", p.Activities[0].HRZonesS)
			}
		})
	}
}
