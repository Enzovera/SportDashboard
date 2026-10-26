package dashboard

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"forma/api/internal/metrics"
	"forma/api/internal/plan"
	"forma/api/internal/projector"
)

const (
	blockedDays        = 7  // blocked proposals shown
	historyDecidedDays = 30 // decisions shown
)

var blockLabel = map[string]string{"base": "Baza", "build": "Budowa", "peak": "Szczyt", "taper": "Taper"}

// dayTypeOf: the plan says it where it covers the date; otherwise it is inferred from what was done.
func dayTypeOf(data Data, date, today time.Time) *string {
	if data.Plan != nil && data.Plan.Covers(date) {
		return ptr(data.Plan.Day(date).DayType)
	}
	return inferDayType(data, date, today)
}

// planTargetAt is the plan's nutrition target for the day; nil without a plan target or for a day the plan does
// not cover (before it started, the Fitatu goal of that time stays).
func planTargetAt(data Data, date time.Time, dayType *string) *DietTarget {
	if data.Plan == nil || dayType == nil || !data.Plan.Covers(date) {
		return nil
	}
	t := data.Plan.TargetAt(date, *dayType)
	if t == nil {
		return nil
	}
	return &DietTarget{Kcal: t.Kcal, ProteinG: t.ProteinG, CarbsG: t.CarbsG, FatG: t.FatG}
}

func stamp(t time.Time) string { return t.In(projector.Warsaw).Format(time.RFC3339) }

// effectiveFrom of a pending proposal: as proposed, else tomorrow (the earliest a decision can matter).
func effectiveFrom(p plan.Proposal, today time.Time) time.Time {
	if p.EffectiveFrom != nil {
		return p.EffectiveFrom.Time
	}
	return today.Add(day)
}

// touches reports whether a proposal changes the given day of the plan.
func touches(st *plan.State, p plan.Proposal, date, today time.Time) bool {
	for _, d := range p.Change.Days {
		if d.Date.Equal(date) {
			return true
		}
	}
	if len(p.Change.Targets) == 0 || date.Before(effectiveFrom(p, today)) {
		return false
	}
	dt := st.Day(date).DayType
	for _, t := range p.Change.Targets {
		if t.DayType == dt {
			return true
		}
	}
	return false
}

// ---------- /api/plan ----------

func BuildPlan(data Data, today time.Time) PlanResponse {
	resp := PlanResponse{Date: iso(today)}
	st := data.Plan
	if st == nil {
		return resp
	}
	resp.ProposalsPending = ptr(len(st.Pending()))
	race := st.Goal.Date.Time
	pv := &PlanView{
		Goal: RaceGoal{Race: st.Goal.Race, Date: iso(race), WeeksLeft: int(math.Max(0, math.Floor(race.Sub(today).Hours()/24/7))),
			TargetTimeS: st.Goal.TargetTimeS, TargetPaceSPerKm: st.Goal.TargetPaceSPerKm, Secondary: st.Goal.Secondary},
		Blocks: []PlanBlock{},
	}
	if pv.Goal.Secondary == nil {
		pv.Goal.Secondary = []plan.Secondary{}
	}
	for _, b := range st.Blocks {
		weeks := int(math.Round((b.To.Sub(b.From.Time).Hours()/24 + 1) / 7))
		pv.Blocks = append(pv.Blocks, PlanBlock{Kind: b.Kind, From: iso(b.From.Time), To: iso(b.To.Time), Weeks: weeks, Summary: b.Summary})
	}
	thisWeek := metrics.WeekStart(today)
	if w := planWeek(data, thisWeek, today); w.PlannedKm > 0 || w.PlannedStrength > 0 {
		pv.CurrentWeek = &w
	}
	if next := thisWeek.Add(7 * day); !next.After(race) {
		w := planWeek(data, next, today)
		pv.NextWeek = &w
	}
	resp.Plan = pv
	return resp
}

func planWeek(data Data, start, today time.Time) PlanWeek {
	st := data.Plan
	w := PlanWeek{From: iso(start), To: iso(start.Add(6 * day)), Days: []PlanDay{}}
	if b, n := st.BlockAt(start); b != nil {
		w.Block, w.WeekInBlock = ptr(b.Kind), ptr(n)
	}
	km := runKmByDay(data.Activities)
	sets := map[time.Time]int{}
	for _, wo := range data.Workouts {
		sets[wo.LocalDate] += wo.WorkingSets
	}
	started := !start.After(today)
	if started {
		end := start.Add(6 * day)
		if end.After(today) {
			end = today
		}
		w.DoneKm = ptr(sumKm(km, start, end))
		done := 0
		for _, wo := range data.Workouts {
			if !wo.LocalDate.Before(start) && !wo.LocalDate.After(end) {
				done++
			}
		}
		w.DoneStrength = ptr(done)
	}
	pending := st.Pending()
	for _, date := range metrics.Days(start, start.Add(6*day)) {
		pd := st.Day(date)
		day := PlanDay{Date: iso(date), Sessions: []PlanSession{}, DayType: pd.DayType, Proposals: []DayProposal{},
			Target: macros(planTargetAt(data, date, &pd.DayType))}
		for _, s := range pd.Sessions {
			day.Sessions = append(day.Sessions, PlanSession{Kind: s.Kind, Title: s.Title, Detail: s.Detail, DistanceKm: s.DistanceKm})
			if s.Kind == "strength" {
				w.PlannedStrength++
			}
		}
		if r := pd.RunKm(); r > 0 {
			day.RunKm = ptr(round1(r))
			w.PlannedKm += r
		}
		if started && date.Before(today) {
			dd := &DayDone{}
			if v := km[date]; v > 0 {
				dd.RunKm = ptr(round1(v))
			}
			if v := sets[date]; v > 0 {
				dd.StrengthSets = ptr(v)
			}
			day.Done = dd
		}
		for i := len(pending) - 1; i >= 0; i-- { // oldest first
			if p := pending[i]; touches(st, p, date, today) {
				day.Proposals = append(day.Proposals, DayProposal{ID: p.ProposalID, Summary: p.Summary})
			}
		}
		w.Days = append(w.Days, day)
	}
	w.PlannedKm = round1(w.PlannedKm)
	return w
}

func macros(t *DietTarget) *Macros {
	if t == nil {
		return nil
	}
	return &Macros{Kcal: t.Kcal, ProteinG: t.ProteinG, CarbsG: t.CarbsG, FatG: t.FatG}
}

// ---------- /api/proposals ----------

func BuildProposals(data Data, blocked []plan.BlockedProposal, today time.Time) ProposalsResponse {
	resp := ProposalsResponse{Date: iso(today), Pending: []ProposalView{}, Blocked: []BlockedView{}, History: []ProposalDecision{}}
	for _, b := range blocked {
		if b.ProposedAt.Before(today.Add(-blockedDays * day)) {
			continue
		}
		resp.Blocked = append(resp.Blocked, BlockedView{ID: b.ID, Kind: b.Kind, Origin: b.Origin, ProposedAt: stamp(b.ProposedAt),
			Title: b.Title, Checks: b.Checks})
	}
	st := data.Plan
	if st == nil {
		return resp
	}
	pending := st.Pending()
	for _, p := range pending {
		v := ProposalView{ID: p.ProposalID, Kind: p.Kind, Origin: p.Origin, ProposedAt: stamp(p.ProposedAt), Title: p.Title,
			Rationale: p.Rationale, Diff: p.Diff, Checks: p.Checks}
		if p.EffectiveFrom != nil {
			v.EffectiveFrom = ptr(iso(p.EffectiveFrom.Time))
		}
		if v.Checks == nil {
			v.Checks = []plan.Check{}
		}
		resp.Pending = append(resp.Pending, v)
	}
	for _, p := range st.Proposals {
		if p.Decision != nil && !p.Decision.At.Before(today.Add(-historyDecidedDays*day)) {
			resp.History = append(resp.History, decisionView(p))
		}
	}
	sort.SliceStable(resp.History, func(i, j int) bool { return resp.History[i].DecidedAt > resp.History[j].DecidedAt })
	resp.Week = weekDiff(data, pending, today)
	return resp
}

func decisionView(p plan.Proposal) ProposalDecision {
	dec := "rejected"
	if p.Decision.Accepted {
		dec = "accepted"
	}
	return ProposalDecision{ID: p.ProposalID, Kind: p.Kind, Origin: p.Origin, Title: p.Title, ProposedAt: stamp(p.ProposedAt),
		DecidedAt: stamp(p.Decision.At), Decision: dec, Comment: p.Decision.Comment}
}

// weekDiff shows the week first touched by pending proposals (from today on): the plan now and, per day,
// what each pending proposal would make of it.
func weekDiff(data Data, pending []plan.Proposal, today time.Time) *WeekDiff {
	st := data.Plan
	if len(pending) == 0 {
		return nil
	}
	var first time.Time
	for _, p := range pending {
		dates := []time.Time{}
		for _, d := range p.Change.Days {
			dates = append(dates, d.Date.Time)
		}
		if len(p.Change.Targets) > 0 {
			dates = append(dates, effectiveFrom(p, today))
		}
		for _, d := range dates {
			if !d.Before(today) && (first.IsZero() || d.Before(first)) {
				first = d
			}
		}
	}
	if first.IsZero() {
		return nil // only reminder changes, or nothing ahead
	}
	start := metrics.WeekStart(first)
	wd := &WeekDiff{From: iso(start), To: iso(start.Add(6 * day)), Days: []WeekDiffDay{}}
	if b, n := st.BlockAt(start); b != nil {
		wd.Block = ptr(blockLabel[b.Kind] + " · tydz. " + strconv.Itoa(n))
	}
	previews := make([]*plan.State, len(pending))
	for i, p := range pending {
		previews[i] = st.With(p.Change, effectiveFrom(p, today))
	}
	for _, date := range metrics.Days(start, start.Add(6*day)) {
		now := dayPlan(st, date)
		row := WeekDiffDay{Date: iso(date), Now: now, Changes: []DayChange{}}
		for i := len(pending) - 1; i >= 0; i-- { // oldest first: later entries win
			if after := dayPlan(previews[i], date); !after.equal(now) {
				row.Changes = append(row.Changes, DayChange{ProposalID: pending[i].ProposalID, After: after})
			}
		}
		wd.Days = append(wd.Days, row)
	}
	return wd
}

func dayPlan(st *plan.State, date time.Time) DayPlan {
	d := st.Day(date)
	var parts []string
	for _, s := range d.Sessions {
		switch {
		case s.Short != nil:
			parts = append(parts, *s.Short)
		case s.Kind == "run" && s.DistanceKm != nil:
			parts = append(parts, s.Title+" "+kmLabel(*s.DistanceKm)+" km")
		default:
			parts = append(parts, s.Title)
		}
	}
	dp := DayPlan{Session: "Odpoczynek"}
	if len(parts) > 0 {
		dp.Session = strings.Join(parts, " + ")
	}
	if km := d.RunKm(); km > 0 {
		dp.RunKm = ptr(round1(km))
	}
	if t := st.TargetAt(date, d.DayType); t != nil {
		dp.Kcal, dp.CarbsG = t.Kcal, t.CarbsG
	}
	return dp
}

func (a DayPlan) equal(b DayPlan) bool {
	return a.Session == b.Session && a.Kcal == b.Kcal && a.CarbsG == b.CarbsG && eqp(a.RunKm, b.RunKm)
}

func kmLabel(km float64) string {
	if km == math.Trunc(km) {
		return strconv.Itoa(int(km))
	}
	return strings.Replace(strconv.FormatFloat(km, 'f', 1, 64), ".", ",", 1)
}

func eqp(a, b *float64) bool { return (a == nil && b == nil) || (a != nil && b != nil && *a == *b) }

// planTargets are the targets per day type in force today and the pending proposals that change them.
func planTargets(data Data, today time.Time) ([]DayTarget, []TargetProposal) {
	targets, proposals := []DayTarget{}, []TargetProposal{}
	st := data.Plan
	if st == nil {
		return targets, proposals
	}
	for _, dt := range plan.DayTypes {
		if t := st.TargetAt(today, dt); t != nil {
			targets = append(targets, DayTarget{DayType: dt, Kcal: t.Kcal, ProteinG: t.ProteinG, CarbsG: t.CarbsG, FatG: t.FatG})
		}
	}
	for _, p := range st.Pending() {
		for _, t := range p.Change.Targets {
			for _, f := range []struct {
				name string
				v    *float64
			}{{"kcal", t.Kcal}, {"proteinG", t.ProteinG}, {"carbsG", t.CarbsG}, {"fatG", t.FatG}} {
				if f.v != nil {
					proposals = append(proposals, TargetProposal{ID: p.ProposalID, DayType: t.DayType, Field: f.name, Value: *f.v})
				}
			}
		}
	}
	return targets, proposals
}

// reminderProposal is the pending coach proposal of a reminder time, if any (newest).
func reminderProposal(data Data) *ReminderProposal {
	if data.Plan == nil {
		return nil
	}
	for _, p := range data.Plan.Pending() {
		if r := p.Change.Reminder; r != nil {
			return &ReminderProposal{ID: p.ProposalID, SupplementID: r.SupplementID, SupplementName: r.SupplementName,
				Text: p.Rationale, CurrentTime: r.From, ProposedTime: r.To}
		}
	}
	return nil
}
