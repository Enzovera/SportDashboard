// Package dashboard serves the views of the dashboard from projections. Builders are pure functions of
// (Data, today, now) so every number the UI shows is unit-tested; Service only adds loading and the clock.
package dashboard

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"forma/api/internal/metrics"
	"forma/api/internal/projector"
)

const (
	day = 24 * time.Hour
	// History loaded for every request: CTL (42-day constant) needs a few months to settle.
	historyDays = 180
	// A source without a successful run for this long is shown as stale.
	staleAfter = 3 * time.Hour
	// HRV 7-day mean at or above this share of the 60-day baseline counts as normal.
	hrvNormalShare = 0.95
)

// Sources always reported; others (fitatu, later) appear once they have runs.
var coreSources = []string{"garmin", "hevy"}

type Service struct {
	store Store
	plans PlanStore
	now   func() time.Time
}

func NewService(store Store, plans PlanStore) *Service {
	return &Service{store: store, plans: plans, now: time.Now}
}

// load reads projections from `since` and the current plan.
func (s *Service) load(ctx context.Context, since time.Time) (Data, error) {
	data, err := s.store.Load(ctx, since)
	if err != nil {
		return data, err
	}
	data.Plan, err = s.plans.Current(ctx)
	return data, err
}

// Today returns the local calendar date (UTC midnight) for now.
func Today(now time.Time) time.Time {
	l := now.In(projector.Warsaw)
	return time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) Today(ctx context.Context) (TodayResponse, error) {
	now := s.now()
	today := Today(now)
	data, err := s.load(ctx, today.Add(-historyDays*day))
	if err != nil {
		return TodayResponse{}, err
	}
	return BuildToday(data, today, now), nil
}

func (s *Service) Training(ctx context.Context, rng string) (TrainingResponse, error) {
	now := s.now()
	today := Today(now)
	data, err := s.load(ctx, today.Add(-historyDays*day))
	if err != nil {
		return TrainingResponse{}, err
	}
	return BuildTraining(data, today, rng, now), nil
}

func (s *Service) Body(ctx context.Context) (BodyResponse, error) {
	now := s.now()
	today := Today(now)
	data, err := s.load(ctx, today.Add(-historyDays*day))
	if err != nil {
		return BodyResponse{}, err
	}
	return BuildBody(data, today, now), nil
}

func (s *Service) Nutrition(ctx context.Context) (NutritionResponse, error) {
	now := s.now()
	today := Today(now)
	data, err := s.load(ctx, today.Add(-historyDays*day))
	if err != nil {
		return NutritionResponse{}, err
	}
	sd, err := s.store.LoadSupplements(ctx, today)
	if err != nil {
		return NutritionResponse{}, err
	}
	resp := BuildNutrition(data, today, now)
	resp.Supplements = SupplementsSummary(BuildSupplements(data, sd, today))
	return resp, nil
}

func (s *Service) Supplements(ctx context.Context) (SupplementsResponse, error) {
	today := Today(s.now())
	data, sd, err := s.loadSupplements(ctx, today)
	if err != nil {
		return SupplementsResponse{}, err
	}
	return BuildSupplements(data, sd, today), nil
}

// LogSupplement appends SupplementTaken / SupplementSkipped for today and returns the updated checklist item.
// Only today can be ticked: the date guards against a page left open past midnight.
func (s *Service) LogSupplement(ctx context.Context, id, status, date string) (SupplementItem, error) {
	today := Today(s.now())
	typ, ok := map[string]string{"taken": SupplementTaken, "skipped": SupplementSkipped}[status]
	if !ok || date != iso(today) {
		return SupplementItem{}, fmt.Errorf("%w: status must be taken or skipped and date today (%s)", ErrInvalid, iso(today))
	}
	if err := s.store.AppendSupplementLog(ctx, id, today, typ); err != nil {
		return SupplementItem{}, err
	}
	data, sd, err := s.loadSupplements(ctx, today)
	if err != nil {
		return SupplementItem{}, err
	}
	for _, it := range BuildSupplements(data, sd, today).Checklist {
		if it.ID == id {
			return it, nil
		}
	}
	return SupplementItem{}, ErrNotFound // logged, but the supplement is not active today
}

func (s *Service) loadSupplements(ctx context.Context, today time.Time) (Data, SupplementData, error) {
	data, err := s.load(ctx, today.Add(-regularityDays*day))
	if err != nil {
		return Data{}, SupplementData{}, err
	}
	sd, err := s.store.LoadSupplements(ctx, today.Add(-regularityDays*day))
	return data, sd, err
}

// ---------- shared helpers ----------

func iso(t time.Time) string { return t.Format(time.DateOnly) }

func ptr[T any](v T) *T { return &v }

func round1(v float64) float64 { return metrics.Round(v, 1) }

func isRun(sport string) bool { return projector.Activity{Sport: sport}.IsRun() }

// BuildSync reports every core source plus any other with runs; also used by the MCP server.
func BuildSync(rows []SyncRow, now time.Time) []SyncStatus {
	bySource := map[string]SyncRow{}
	for _, r := range rows {
		bySource[r.Source] = r
	}
	sources := append([]string{}, coreSources...)
	for _, r := range rows {
		if !contains(sources, r.Source) {
			sources = append(sources, r.Source)
		}
	}
	out := make([]SyncStatus, 0, len(sources))
	for _, src := range sources {
		r, ok := bySource[src]
		st := SyncStatus{Source: src, Status: "stale"}
		if ok && r.LastSuccessAt != nil {
			st.LastSuccessAt = ptr(r.LastSuccessAt.In(projector.Warsaw).Format(time.RFC3339))
		}
		switch {
		case !ok:
		case r.LastStatus == "error" || r.LastStatus == "partial":
			st.Status = r.LastStatus
		case r.LastSuccessAt != nil && now.Sub(*r.LastSuccessAt) <= staleAfter:
			st.Status = "ok"
		}
		out = append(out, st)
	}
	return out
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// runKmByDay sums running distance per local date, in km.
func runKmByDay(acts []ActivityRow) map[time.Time]float64 {
	out := map[time.Time]float64{}
	for _, a := range acts {
		if isRun(a.Sport) && a.DistanceM != nil {
			out[a.LocalDate] += *a.DistanceM / 1000
		}
	}
	return out
}

// loadSeries covers the whole loaded history, so any displayed range has points even before the first activity.
func loadSeries(data Data, to time.Time) []metrics.LoadPoint {
	daily := map[time.Time]float64{}
	for _, a := range data.Activities {
		if a.Load != nil {
			daily[a.LocalDate] += *a.Load
		}
	}
	return metrics.LoadSeries(daily, to.Add(-historyDays*day), to)
}

func sumKm(byDay map[time.Time]float64, from, to time.Time) float64 {
	var km float64
	for _, d := range metrics.Days(from, to) {
		km += byDay[d]
	}
	return round1(km)
}

// ---------- /api/today ----------

func BuildToday(data Data, today, now time.Time) TodayResponse {
	series := loadSeries(data, today)
	resp := TodayResponse{
		Date:      iso(today),
		Sync:      BuildSync(data.Sync, now),
		Week:      buildWeek(data, today),
		Readiness: buildReadiness(data, today, series),
		Weight:    buildWeight(data.Days, today),
	}
	if p := BuildPlan(data, today); p.Plan != nil {
		resp.RaceWeeksLeft = ptr(p.Plan.Goal.WeeksLeft)
		resp.ProposalsPending = p.ProposalsPending
	}
	return resp
}

func buildWeek(data Data, today time.Time) WeekGrid {
	from := metrics.WeekStart(today)
	to := from.Add(6 * day)
	km := runKmByDay(data.Activities)
	strength := map[time.Time]string{}
	done := 0
	for _, w := range data.Workouts {
		if w.LocalDate.Before(from) || w.LocalDate.After(to) {
			continue
		}
		done++
		if _, ok := strength[w.LocalDate]; !ok {
			title := "Siła"
			if w.Title != nil && *w.Title != "" {
				title = *w.Title
			}
			strength[w.LocalDate] = title
		}
	}
	g := WeekGrid{From: iso(from), To: iso(to), RunKmDone: sumKm(km, from, to), StrengthDone: done}
	for _, d := range metrics.Days(from, to) {
		wd := WeekDay{Date: iso(d)}
		if v := km[d]; v > 0 {
			wd.RunKm = ptr(round1(v))
		}
		if t, ok := strength[d]; ok {
			wd.Strength = ptr(t)
		}
		g.Days = append(g.Days, wd)
	}
	return g
}

// window returns values of field for days in [from, to].
func window(days []DayRow, from, to time.Time, field func(DayRow) *float64) []*float64 {
	var out []*float64
	for _, d := range days {
		if !d.Date.Before(from) && !d.Date.After(to) {
			out = append(out, field(d))
		}
	}
	return out
}

func intAsFloat(p *int) *float64 {
	if p == nil {
		return nil
	}
	return ptr(float64(*p))
}

func buildReadiness(data Data, today time.Time, series []metrics.LoadPoint) Readiness {
	r := Readiness{HRVTone: "neutral"}
	var todayRow *DayRow
	for i := range data.Days {
		if data.Days[i].Date.Equal(today) {
			todayRow = &data.Days[i]
		}
	}
	hrv := func(d DayRow) *float64 { return d.HRVMs }
	sleep := func(d DayRow) *float64 { return intAsFloat(d.SleepS) }
	rhr := func(d DayRow) *float64 { return intAsFloat(d.RestingHR) }

	// HRV: 7-day mean against a 60-day baseline, never a single night (docs/design.md, "Gotowość").
	if v, ok := metrics.Mean(window(data.Days, today.Add(-6*day), today, hrv), 3); ok {
		r.HRV7dMs = ptr(math.Round(v))
	}
	if v, ok := metrics.Mean(window(data.Days, today.Add(-59*day), today, hrv), 14); ok {
		r.HRVBaselineMs = ptr(math.Round(v))
	}
	if r.HRV7dMs != nil && r.HRVBaselineMs != nil {
		if *r.HRV7dMs >= *r.HRVBaselineMs*hrvNormalShare {
			r.HRVTone, r.HRVLabel = "good", ptr("w normie")
		} else {
			r.HRVTone, r.HRVLabel = "warn", ptr("poniżej bazy")
		}
	}

	if todayRow != nil {
		r.SleepS = todayRow.SleepS
		r.RestingHR = todayRow.RestingHR
	}
	if v, ok := metrics.Mean(window(data.Days, today.Add(-6*day), today, sleep), 3); ok {
		r.SleepAvg7dS = ptr(int(math.Round(v)))
	}
	if r.RestingHR != nil {
		if v, ok := metrics.Mean(window(data.Days, today.Add(-7*day), today.Add(-day), rhr), 3); ok {
			r.RestingHRDelta7d = ptr(*r.RestingHR - int(math.Round(v)))
		}
	}
	if tsb, ok := metrics.TSB(series, today); ok {
		r.TSB = ptr(math.Round(tsb))
		r.Form = ptr(string(metrics.FormOf(tsb)))
	}
	return r
}

// buildWeight: 7-day mean once there are at least 3 weigh-ins, and its change against the week before.
func buildWeight(days []DayRow, today time.Time) *WeightTrend {
	w := func(d DayRow) *float64 { return d.WeightKg }
	cur, ok := metrics.Mean(window(days, today.Add(-6*day), today, w), 3)
	if !ok {
		return nil
	}
	t := &WeightTrend{Avg7dKg: round1(cur)}
	if prev, ok := metrics.Mean(window(days, today.Add(-13*day), today.Add(-7*day), w), 3); ok {
		t.DeltaPerWeekKg = ptr(round1(cur - prev))
	}
	return t
}

// ---------- /api/training ----------

func rangeWeeks(rng string) (string, int) {
	switch rng {
	case "4w":
		return "4w", 4
	case "block":
		// No plan blocks yet (stage 4): a block reads as the last 12 weeks.
		return "block", 12
	default:
		return "12w", 12
	}
}

func BuildTraining(data Data, today time.Time, rng string, now time.Time) TrainingResponse {
	rng, weeks := rangeWeeks(rng)
	thisWeek := metrics.WeekStart(today)
	from := thisWeek.Add(-time.Duration(weeks-1) * 7 * day)
	km := runKmByDay(data.Activities)
	series := loadSeries(data, today)

	resp := TrainingResponse{
		Range:            rng,
		From:             iso(from),
		To:               iso(today),
		Sync:             BuildSync(data.Sync, now),
		Load:             buildLoad(km, series, thisWeek, today, from),
		RecentRuns:       recentRuns(data.Activities, 5),
		SetsPerMuscle:    setsPerMuscle(data.Workouts, thisWeek),
		StrengthSessions: strengthSessions(data.Workouts, thisWeek),
	}
	for ws := from; !ws.After(thisWeek); ws = ws.Add(7 * day) {
		resp.WeeklyKm = append(resp.WeeklyKm, WeeklyKm{WeekStart: iso(ws), Km: sumKm(km, ws, ws.Add(6*day))})
	}
	for _, p := range series {
		if !p.Date.Before(from) {
			resp.LoadSeries = append(resp.LoadSeries, LoadPoint{Date: iso(p.Date), CTL: round1(p.CTL), ATL: round1(p.ATL)})
		}
	}
	return resp
}

func buildLoad(km map[time.Time]float64, series []metrics.LoadPoint, thisWeek, today, from time.Time) LoadSummary {
	l := LoadSummary{
		WeekFrom:   iso(thisWeek),
		WeekTo:     iso(thisWeek.Add(6 * day)),
		WeekKm:     sumKm(km, thisWeek, thisWeek.Add(6*day)),
		PrevWeekKm: ptr(sumKm(km, thisWeek.Add(-7*day), thisWeek.Add(-day))),
	}
	if p, ok := metrics.At(series, today); ok {
		l.CTL, l.ATL = ptr(math.Round(p.CTL)), ptr(math.Round(p.ATL))
		if start, ok := metrics.At(series, from.Add(-day)); ok {
			l.CTLDeltaRange = ptr(math.Round(p.CTL - start.CTL))
		}
		if weekAgo, ok := metrics.At(series, today.Add(-7*day)); ok {
			l.ATLDeltaWeek = ptr(math.Round(p.ATL - weekAgo.ATL))
		}
	}
	if tsb, ok := metrics.TSB(series, today); ok {
		l.TSB = ptr(math.Round(tsb))
		l.Form = ptr(string(metrics.FormOf(tsb)))
	}
	return l
}

func recentRuns(acts []ActivityRow, n int) []RunRow {
	out := []RunRow{}
	for i := len(acts) - 1; i >= 0 && len(out) < n; i-- {
		a := acts[i]
		if !isRun(a.Sport) || a.DistanceM == nil || *a.DistanceM <= 0 {
			continue
		}
		r := RunRow{ID: a.ID, Date: iso(a.LocalDate), Title: "Bieg", DistanceKm: round1(*a.DistanceM / 1000), AvgHR: a.AvgHR}
		if a.Name != nil && *a.Name != "" {
			r.Title = *a.Name
		}
		if a.DurationS != nil {
			r.PaceSPerKm = ptr(math.Round(*a.DurationS / (*a.DistanceM / 1000)))
		}
		if a.Load != nil {
			r.Load = ptr(math.Round(*a.Load))
		}
		out = append(out, r)
	}
	return out
}

func weekWorkouts(ws []WorkoutRow, weekStart time.Time) []WorkoutRow {
	var out []WorkoutRow
	for _, w := range ws {
		if !w.LocalDate.Before(weekStart) && w.LocalDate.Before(weekStart.Add(7*day)) {
			out = append(out, w)
		}
	}
	return out
}

func setsPerMuscle(ws []WorkoutRow, weekStart time.Time) SetsPerMuscle {
	counts := map[string]int{}
	for _, w := range weekWorkouts(ws, weekStart) {
		for m, n := range w.Muscles {
			if m != "cardio" {
				counts[m] += n
			}
		}
	}
	items := []MuscleSets{}
	for m, n := range counts {
		items = append(items, MuscleSets{Muscle: m, Sets: n})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Sets != items[j].Sets {
			return items[i].Sets > items[j].Sets
		}
		return items[i].Muscle < items[j].Muscle
	})
	return SetsPerMuscle{WeekFrom: iso(weekStart), WeekTo: iso(weekStart.Add(6 * day)), Items: items}
}

func strengthSessions(ws []WorkoutRow, weekStart time.Time) []StrengthRow {
	out := []StrengthRow{}
	for _, w := range weekWorkouts(ws, weekStart) {
		r := StrengthRow{ID: w.ID, Date: iso(w.LocalDate), Title: "Siła", Sets: w.WorkingSets, TonnageKg: math.Round(w.TonnageKg)}
		if w.Title != nil && *w.Title != "" {
			r.Title = *w.Title
		}
		if w.EndedAt != nil {
			r.DurationMin = int(math.Round(w.EndedAt.Sub(w.StartedAt).Minutes()))
		}
		out = append(out, r)
	}
	return out
}

// ---------- plan ----------

func (s *Service) Plan(ctx context.Context) (PlanResponse, error) {
	today := Today(s.now())
	data, err := s.load(ctx, metrics.WeekStart(today))
	if err != nil {
		return PlanResponse{}, err
	}
	return BuildPlan(data, today), nil
}

func (s *Service) Proposals(ctx context.Context) (ProposalsResponse, error) {
	today := Today(s.now())
	st, err := s.plans.Current(ctx)
	if err != nil {
		return ProposalsResponse{}, err
	}
	blocked, err := s.plans.Blocked(ctx, today.Add(-blockedDays*day))
	if err != nil {
		return ProposalsResponse{}, err
	}
	return BuildProposals(Data{Plan: st}, blocked, today), nil
}

// Decide records the human decision on a proposal; only the dashboard calls it (docs/design.md, "Stan trenera").
func (s *Service) Decide(ctx context.Context, id string, accept bool, comment *string) (ProposalDecision, error) {
	if comment != nil {
		if c := strings.TrimSpace(*comment); c == "" {
			comment = nil
		} else {
			comment = &c
		}
	}
	p, err := s.plans.Decide(ctx, id, accept, comment, s.now())
	if err != nil {
		return ProposalDecision{}, err
	}
	return decisionView(p), nil
}
