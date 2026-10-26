package mcpserver

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"forma/api/internal/dashboard"
	"forma/api/internal/metrics"
	"forma/api/internal/projector"
)

// Tool outputs. Builders are pure functions of (Data, range, today) like the dashboard views, so every number
// the coach sees is computed and tested here; the model only interprets them (docs/design.md, "Metryki").
// Nullable values are pointers: null means "no data", never zero.

const (
	day = 24 * time.Hour
	// History loaded before a range so ATL/CTL have settled on its first day.
	historyDays = 180
	// Longest range a single call may ask for.
	maxRangeDays = 366
	// Default range when `from` is omitted.
	defaultRangeDays = 28
)

// ---------- input parsing ----------

func parseDate(s string, def time.Time) (time.Time, error) {
	if s == "" {
		return def, nil
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("date %q: expected YYYY-MM-DD", s)
	}
	return t, nil
}

// parseRange defaults `to` to today and `from` to the 28 days ending on `to`.
func parseRange(from, to string, today time.Time) (time.Time, time.Time, error) {
	t, err := parseDate(to, today)
	if err != nil {
		return t, t, err
	}
	f, err := parseDate(from, t.Add(-(defaultRangeDays-1)*day))
	if err != nil {
		return f, t, err
	}
	if f.After(t) {
		return f, t, fmt.Errorf("from %s is after to %s", iso(f), iso(t))
	}
	if days := int(t.Sub(f)/day) + 1; days > maxRangeDays {
		return f, t, fmt.Errorf("range of %d days exceeds %d", days, maxRangeDays)
	}
	return f, t, nil
}

func inRange(d, from, to time.Time) bool { return !d.Before(from) && !d.After(to) }

func iso(t time.Time) string { return t.Format(time.DateOnly) }

func ptr[T any](v T) *T { return &v }

func r1(v float64) float64 { return metrics.Round(v, 1) }

func r1p(v float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	return ptr(r1(v))
}

func intAsFloat(p *int) *float64 {
	if p == nil {
		return nil
	}
	return ptr(float64(*p))
}

// ---------- shared shapes ----------

type Activity struct {
	Date         string   `json:"date"`
	StartedAt    string   `json:"started_at"` // local time, RFC 3339
	Sport        string   `json:"sport"`
	Name         *string  `json:"name"`
	DistanceKm   *float64 `json:"distance_km"`
	DurationMin  *float64 `json:"duration_min"`
	PaceMinPerKm *string  `json:"pace_min_per_km"` // runs only, "m:ss"
	AvgHR        *int     `json:"avg_hr"`
	TrainingLoad *float64 `json:"training_load"`
}

type Workout struct {
	Date         string         `json:"date"`
	Title        *string        `json:"title"`
	DurationMin  *float64       `json:"duration_min"`
	WorkingSets  int            `json:"working_sets"`
	TonnageKg    float64        `json:"tonnage_kg"`
	SetsByMuscle map[string]int `json:"sets_by_muscle"`
}

type LoadState struct {
	ATL  *float64 `json:"atl"`  // end of day
	CTL  *float64 `json:"ctl"`  // end of day
	TSB  *float64 `json:"tsb"`  // form going into the day: yesterday's CTL − ATL
	Form *string  `json:"form"` // fresh | neutral | tired | overreached
}

func pace(distM, durS float64) string {
	sPerKm := int(math.Round(durS / (distM / 1000)))
	return fmt.Sprintf("%d:%02d", sPerKm/60, sPerKm%60)
}

func toActivity(a dashboard.ActivityRow) Activity {
	out := Activity{
		Date:         iso(a.LocalDate),
		StartedAt:    a.StartedAt.In(projector.Warsaw).Format(time.RFC3339),
		Sport:        a.Sport,
		Name:         a.Name,
		AvgHR:        a.AvgHR,
		TrainingLoad: a.Load,
	}
	if a.DistanceM != nil && *a.DistanceM > 0 {
		out.DistanceKm = ptr(metrics.Round(*a.DistanceM/1000, 2))
	}
	if a.DurationS != nil {
		out.DurationMin = ptr(r1(*a.DurationS / 60))
	}
	if (projector.Activity{Sport: a.Sport}).IsRun() && a.DistanceM != nil && *a.DistanceM > 0 && a.DurationS != nil {
		out.PaceMinPerKm = ptr(pace(*a.DistanceM, *a.DurationS))
	}
	return out
}

func toWorkout(w dashboard.WorkoutRow) Workout {
	out := Workout{
		Date:         iso(w.LocalDate),
		Title:        w.Title,
		WorkingSets:  w.WorkingSets,
		TonnageKg:    r1(w.TonnageKg),
		SetsByMuscle: w.Muscles,
	}
	if out.SetsByMuscle == nil {
		out.SetsByMuscle = map[string]int{}
	}
	if w.EndedAt != nil {
		out.DurationMin = ptr(r1(w.EndedAt.Sub(w.StartedAt).Minutes()))
	}
	return out
}

func loadSeries(data dashboard.Data, to time.Time) []metrics.LoadPoint {
	daily := dailyLoad(data.Activities)
	return metrics.LoadSeries(daily, to.Add(-historyDays*day), to)
}

func dailyLoad(acts []dashboard.ActivityRow) map[time.Time]float64 {
	daily := map[time.Time]float64{}
	for _, a := range acts {
		if a.Load != nil {
			daily[a.LocalDate] += *a.Load
		}
	}
	return daily
}

func loadAt(series []metrics.LoadPoint, d time.Time) LoadState {
	var s LoadState
	if p, ok := metrics.At(series, d); ok {
		s.ATL, s.CTL = ptr(r1(p.ATL)), ptr(r1(p.CTL))
	}
	if tsb, ok := metrics.TSB(series, d); ok {
		s.TSB = ptr(r1(tsb))
		s.Form = ptr(string(metrics.FormOf(tsb)))
	}
	return s
}

// dayWindow returns field values for days in [from, to].
func dayWindow(days []dashboard.DayRow, from, to time.Time, field func(dashboard.DayRow) *float64) []*float64 {
	var out []*float64
	for _, d := range days {
		if inRange(d.Date, from, to) {
			out = append(out, field(d))
		}
	}
	return out
}

var (
	fieldHRV    = func(d dashboard.DayRow) *float64 { return d.HRVMs }
	fieldWeight = func(d dashboard.DayRow) *float64 { return d.WeightKg }
	fieldRHR    = func(d dashboard.DayRow) *float64 { return intAsFloat(d.RestingHR) }
	fieldSleepH = func(d dashboard.DayRow) *float64 {
		if d.SleepS == nil {
			return nil
		}
		return ptr(float64(*d.SleepS) / 3600)
	}
)

// Same windows as the Today view (dashboard.buildReadiness, buildWeight).
func hrv7d(days []dashboard.DayRow, d time.Time) *float64 {
	return r1p(metrics.Mean(dayWindow(days, d.Add(-6*day), d, fieldHRV), 3))
}

func hrvBaseline(days []dashboard.DayRow, d time.Time) *float64 {
	return r1p(metrics.Mean(dayWindow(days, d.Add(-59*day), d, fieldHRV), 14))
}

func weight7d(days []dashboard.DayRow, d time.Time) *float64 {
	return r1p(metrics.Mean(dayWindow(days, d.Add(-6*day), d, fieldWeight), 3))
}

// ---------- get_day_summary ----------

type DaySummary struct {
	Date          string                 `json:"date"`
	SleepH        *float64               `json:"sleep_h"`
	SleepScore    *int                   `json:"sleep_score"`
	HRVMs         *int                   `json:"hrv_last_night_ms"`
	HRV7dMs       *float64               `json:"hrv_7d_ms"`
	HRVBaselineMs *float64               `json:"hrv_60d_baseline_ms"`
	HRVStatus     *string                `json:"hrv_status_garmin"`
	RestingHR     *int                   `json:"resting_hr"`
	Steps         *int                   `json:"steps"`
	ActiveKcal    *int                   `json:"active_kcal"`
	TotalKcal     *int                   `json:"total_kcal"`
	WeightKg      *float64               `json:"weight_kg"` // the day's average weigh-in
	Weight7dKg    *float64               `json:"weight_7d_kg"`
	Load          LoadState              `json:"load"`
	Activities    []Activity             `json:"activities"`
	Workouts      []Workout              `json:"workouts"`
	Sync          []dashboard.SyncStatus `json:"sync"`
	Missing       []string               `json:"missing"`
}

func BuildDaySummary(data dashboard.Data, detail *DayDetail, date, now time.Time) DaySummary {
	out := DaySummary{
		Date:       iso(date),
		HRV7dMs:    hrv7d(data.Days, date),
		Weight7dKg: weight7d(data.Days, date),
		Load:       loadAt(loadSeries(data, date), date),
		Activities: []Activity{},
		Workouts:   []Workout{},
		Sync:       dashboard.BuildSync(data.Sync, now),
		// Stage 2 sources: say so instead of letting the model guess.
		Missing: []string{"nutrition (Fitatu not integrated yet)"},
	}
	out.HRVBaselineMs = hrvBaseline(data.Days, date)
	if detail != nil {
		out.SleepScore, out.HRVMs, out.HRVStatus = detail.SleepScore, detail.HRVMs, detail.HRVStatus
		out.RestingHR, out.Steps, out.ActiveKcal, out.TotalKcal = detail.RestingHR, detail.Steps, detail.ActiveKcal, detail.TotalKcal
		if detail.SleepS != nil {
			out.SleepH = ptr(metrics.Round(float64(*detail.SleepS)/3600, 2))
		}
		if detail.WeightKg != nil {
			out.WeightKg = ptr(r1(*detail.WeightKg))
		}
	} else {
		out.Missing = append(out.Missing, "day_summary (no Garmin data for this day)")
	}
	for _, a := range data.Activities {
		if a.LocalDate.Equal(date) {
			out.Activities = append(out.Activities, toActivity(a))
		}
	}
	for _, w := range data.Workouts {
		if w.LocalDate.Equal(date) {
			out.Workouts = append(out.Workouts, toWorkout(w))
		}
	}
	return out
}

// ---------- get_activities ----------

type ActivityList struct {
	From       string     `json:"from"`
	To         string     `json:"to"`
	Activities []Activity `json:"activities"`
}

// BuildActivities lists Garmin activities in the range; sport filters by substring ("running" matches
// treadmill_running and trail_running too).
func BuildActivities(data dashboard.Data, from, to time.Time, sport string) ActivityList {
	out := ActivityList{From: iso(from), To: iso(to), Activities: []Activity{}}
	for _, a := range data.Activities {
		if inRange(a.LocalDate, from, to) && (sport == "" || strings.Contains(a.Sport, sport)) {
			out.Activities = append(out.Activities, toActivity(a))
		}
	}
	return out
}

// ---------- get_strength_volume ----------

var strengthGroups = []any{"week", "muscle", "session"}

type StrengthRow struct {
	Key          string         `json:"key"` // week start (Monday), muscle group or session date
	Title        *string        `json:"title,omitempty"`
	Sessions     int            `json:"sessions"`
	WorkingSets  int            `json:"working_sets"`
	TonnageKg    *float64       `json:"tonnage_kg"` // null for muscle groups: tonnage is not split per muscle
	SetsByMuscle map[string]int `json:"sets_by_muscle,omitempty"`
}

type StrengthVolume struct {
	From    string        `json:"from"`
	To      string        `json:"to"`
	GroupBy string        `json:"group_by"`
	Rows    []StrengthRow `json:"rows"`
}

// BuildStrengthVolume counts Hevy working sets (no warm-ups, reps done) and tonnage, the same as the Training view.
func BuildStrengthVolume(data dashboard.Data, from, to time.Time, groupBy string) StrengthVolume {
	out := StrengthVolume{From: iso(from), To: iso(to), GroupBy: groupBy, Rows: []StrengthRow{}}
	// Weeks are whole ISO weeks, like the weekly metrics of get_trends.
	lo := from
	if groupBy == "week" {
		lo = metrics.WeekStart(from)
	}
	var ws []dashboard.WorkoutRow
	for _, w := range data.Workouts {
		if inRange(w.LocalDate, lo, to) {
			ws = append(ws, w)
		}
	}
	switch groupBy {
	case "session":
		for _, w := range ws {
			out.Rows = append(out.Rows, StrengthRow{Key: iso(w.LocalDate), Title: w.Title, Sessions: 1,
				WorkingSets: w.WorkingSets, TonnageKg: ptr(r1(w.TonnageKg)), SetsByMuscle: w.Muscles})
		}
	case "muscle":
		sets, sessions := map[string]int{}, map[string]int{}
		for _, w := range ws {
			for m, n := range w.Muscles {
				sets[m] += n
				sessions[m]++
			}
		}
		for m, n := range sets {
			out.Rows = append(out.Rows, StrengthRow{Key: m, Sessions: sessions[m], WorkingSets: n})
		}
		sort.Slice(out.Rows, func(i, j int) bool {
			if out.Rows[i].WorkingSets != out.Rows[j].WorkingSets {
				return out.Rows[i].WorkingSets > out.Rows[j].WorkingSets
			}
			return out.Rows[i].Key < out.Rows[j].Key
		})
	default: // week
		for wk := metrics.WeekStart(from); !wk.After(to); wk = wk.Add(7 * day) {
			row := StrengthRow{Key: iso(wk), SetsByMuscle: map[string]int{}}
			var tonnage float64
			for _, w := range ws {
				if inRange(w.LocalDate, wk, wk.Add(6*day)) {
					row.Sessions++
					row.WorkingSets += w.WorkingSets
					tonnage += w.TonnageKg
					for m, n := range w.Muscles {
						row.SetsByMuscle[m] += n
					}
				}
			}
			row.TonnageKg = ptr(r1(tonnage))
			out.Rows = append(out.Rows, row)
		}
	}
	return out
}

// ---------- get_trends ----------

type trendMetric struct {
	unit   string
	weekly bool
	// value for a day (or the week starting on it); ok=false serializes as null.
	value func(c *trendCtx, d time.Time) (float64, bool)
}

type trendCtx struct {
	data   dashboard.Data
	series []metrics.LoadPoint
	days   map[time.Time]dashboard.DayRow
	load   map[time.Time]float64
}

func dayField(field func(dashboard.DayRow) *float64) func(*trendCtx, time.Time) (float64, bool) {
	return func(c *trendCtx, d time.Time) (float64, bool) {
		row, ok := c.days[d]
		if !ok || field(row) == nil {
			return 0, false
		}
		return *field(row), true
	}
}

func weekSum(c *trendCtx, wk time.Time, f func(dashboard.ActivityRow) (float64, bool)) float64 {
	var sum float64
	for _, a := range c.data.Activities {
		if inRange(a.LocalDate, wk, wk.Add(6*day)) {
			if v, ok := f(a); ok {
				sum += v
			}
		}
	}
	return sum
}

func runKm(a dashboard.ActivityRow) (float64, bool) {
	if !(projector.Activity{Sport: a.Sport}).IsRun() || a.DistanceM == nil {
		return 0, false
	}
	return *a.DistanceM / 1000, true
}

var trendMetrics = map[string]trendMetric{
	"training_load": {unit: "garmin load", value: func(c *trendCtx, d time.Time) (float64, bool) { return c.load[d], true }},
	"atl": {unit: "load", value: func(c *trendCtx, d time.Time) (float64, bool) {
		p, ok := metrics.At(c.series, d)
		return p.ATL, ok
	}},
	"ctl": {unit: "load", value: func(c *trendCtx, d time.Time) (float64, bool) {
		p, ok := metrics.At(c.series, d)
		return p.CTL, ok
	}},
	"tsb": {unit: "load", value: func(c *trendCtx, d time.Time) (float64, bool) { return metrics.TSB(c.series, d) }},
	"hrv": {unit: "ms", value: dayField(fieldHRV)},
	"hrv_7d": {unit: "ms", value: func(c *trendCtx, d time.Time) (float64, bool) {
		return metrics.Mean(dayWindow(c.data.Days, d.Add(-6*day), d, fieldHRV), 3)
	}},
	"resting_hr": {unit: "bpm", value: dayField(fieldRHR)},
	"sleep_h":    {unit: "h", value: dayField(fieldSleepH)},
	"weight":     {unit: "kg", value: dayField(fieldWeight)},
	"weight_7d": {unit: "kg", value: func(c *trendCtx, d time.Time) (float64, bool) {
		return metrics.Mean(dayWindow(c.data.Days, d.Add(-6*day), d, fieldWeight), 3)
	}},
	"run_km_week": {unit: "km", weekly: true, value: func(c *trendCtx, wk time.Time) (float64, bool) {
		return weekSum(c, wk, runKm), true
	}},
	"long_run_km_week": {unit: "km", weekly: true, value: func(c *trendCtx, wk time.Time) (float64, bool) {
		var longest float64
		for _, a := range c.data.Activities {
			if km, ok := runKm(a); ok && inRange(a.LocalDate, wk, wk.Add(6*day)) && km > longest {
				longest = km
			}
		}
		return longest, true
	}},
	"strength_sets_week": {unit: "sets", weekly: true, value: func(c *trendCtx, wk time.Time) (float64, bool) {
		var n int
		for _, w := range c.data.Workouts {
			if inRange(w.LocalDate, wk, wk.Add(6*day)) {
				n += w.WorkingSets
			}
		}
		return float64(n), true
	}},
}

// trendNames lists metrics for the tool schema, in a stable order.
func trendNames() []any {
	names := make([]string, 0, len(trendMetrics))
	for n := range trendMetrics {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]any, len(names))
	for i, n := range names {
		out[i] = n
	}
	return out
}

type TrendPoint struct {
	Date  string   `json:"date"` // day, or Monday of the ISO week for weekly metrics
	Value *float64 `json:"value"`
	// Weekly metrics only: change against the week before in %, null when that week was 0 or had no data.
	ChangePct *float64 `json:"change_pct,omitempty"`
}

// changePct is the week-over-week change in %, rounded to whole percent.
func changePct(cur float64, curOK bool, prev float64, prevOK bool) *float64 {
	if !curOK || !prevOK || prev == 0 {
		return nil
	}
	return ptr(math.Round((cur - prev) / prev * 100))
}

type Trend struct {
	Metric      string       `json:"metric"`
	Unit        string       `json:"unit"`
	Granularity string       `json:"granularity"` // day | week
	Points      []TrendPoint `json:"points"`
}

func BuildTrend(data dashboard.Data, metric string, from, to time.Time) (Trend, error) {
	m, ok := trendMetrics[metric]
	if !ok {
		return Trend{}, fmt.Errorf("unknown metric %q", metric)
	}
	c := &trendCtx{data: data, series: loadSeries(data, to), days: map[time.Time]dashboard.DayRow{}, load: dailyLoad(data.Activities)}
	for _, d := range data.Days {
		c.days[d.Date] = d
	}
	out := Trend{Metric: metric, Unit: m.unit, Granularity: "day", Points: []TrendPoint{}}
	step, start := day, from
	if m.weekly {
		out.Granularity, step, start = "week", 7*day, metrics.WeekStart(from)
	}
	for d := start; !d.After(to); d = d.Add(step) {
		v, ok := m.value(c, d)
		p := TrendPoint{Date: iso(d), Value: r1p(v, ok)}
		if m.weekly {
			prev, prevOK := m.value(c, d.Add(-step))
			// Compare rounded values, so the % matches the numbers the model sees.
			p.ChangePct = changePct(r1(v), ok, r1(prev), prevOK)
		}
		out.Points = append(out.Points, p)
	}
	return out, nil
}

// ---------- coach log ----------

var noteKinds = []any{"daily", "weekly", "decision"}

const (
	maxNoteChars = 8000
	// save_coach_note accepts dates this far around today: the weekly run may note next week's plan.
	noteDaysBack, noteDaysAhead = 30, 7
	defaultLogDays, maxLogDays  = 14, 90
)

type NoteOut struct {
	ID        int64  `json:"id"`
	Date      string `json:"date"`
	Kind      string `json:"kind"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type CoachLog struct {
	Since string    `json:"since"`
	Notes []NoteOut `json:"notes"`
}

func toNote(n Note) NoteOut {
	return NoteOut{ID: n.ID, Date: iso(n.Date), Kind: n.Kind, Content: n.Content,
		CreatedAt: n.CreatedAt.In(projector.Warsaw).Format(time.RFC3339)}
}

func logSince(days int, today time.Time) (time.Time, error) {
	if days == 0 {
		days = defaultLogDays
	}
	if days < 1 || days > maxLogDays {
		return time.Time{}, fmt.Errorf("days must be between 1 and %d", maxLogDays)
	}
	return today.Add(-time.Duration(days-1) * day), nil
}

func validateNote(kind, date, content string, today time.Time) (time.Time, string, error) {
	d, err := parseDate(date, today)
	if err != nil {
		return d, "", err
	}
	if d.Before(today.Add(-noteDaysBack*day)) || d.After(today.Add(noteDaysAhead*day)) {
		return d, "", fmt.Errorf("date must be within %d days back and %d days ahead of %s", noteDaysBack, noteDaysAhead, iso(today))
	}
	switch kind {
	case "daily", "weekly", "decision":
	default:
		return d, "", fmt.Errorf("kind must be daily, weekly or decision")
	}
	content = strings.TrimSpace(content)
	if content == "" || len([]rune(content)) > maxNoteChars {
		return d, "", fmt.Errorf("content must have 1 to %d characters", maxNoteChars)
	}
	return d, content, nil
}
