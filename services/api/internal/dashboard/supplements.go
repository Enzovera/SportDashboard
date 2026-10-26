package dashboard

import (
	"math"
	"sort"
	"time"

	"forma/api/internal/metrics"
	"forma/api/internal/projector"
)

const (
	regularityDays = 30
	// Regularity below this share of applicable days is shown as a warning.
	regularityWarnPct = 75
	// Stock that lasts fewer days than this is shown as "zamów" (and pushed, docs/design.md "Suplementacja").
	reorderDays = 14
)

// Supplement log event types (append-only).
const (
	SupplementTaken   = "SupplementTaken"
	SupplementSkipped = "SupplementSkipped"
)

var slotOrder = map[string]int{"morning": 0, "before": 1, "during": 2, "after": 3, "evening": 4}

// ruleApplies tells whether a supplement rule is due on a day of the given type. Without a plan there is no
// session intensity, so a quality run reads as any run day; unknown day type (today before training) = daily only.
func ruleApplies(when string, dayType *string) bool {
	if when == "daily" {
		return true
	}
	if dayType == nil {
		return false
	}
	dt := *dayType
	run := dt == DayRun || dt == DayStrengthRun || dt == DayLongRun
	strength := dt == DayStrength || dt == DayStrengthRun
	switch when {
	case "run_day", "quality_run":
		return run
	case "long_run":
		return dt == DayLongRun
	case "strength_day":
		return strength
	case "training_day":
		return run || strength
	}
	return false
}

func (s SupplementRow) activeOn(d time.Time) bool {
	return !d.Before(s.ActiveFrom) && (s.ActiveTo == nil || !d.After(*s.ActiveTo))
}

// dueRule is the first rule (by slot) that applies on the day; nil when the supplement is not due.
func (s SupplementRow) dueRule(dayType *string) *SupplementRuleRow {
	var best *SupplementRuleRow
	for i := range s.Rules {
		r := &s.Rules[i]
		if ruleApplies(r.When, dayType) && (best == nil || slotOrder[r.Slot] < slotOrder[best.Slot]) {
			best = r
		}
	}
	return best
}

// latestLogs keeps the latest log event per (supplement, date): it is the state of that day.
func latestLogs(logs []SupplementLogRow) map[string]map[time.Time]SupplementLogRow {
	out := map[string]map[time.Time]SupplementLogRow{}
	for _, l := range logs { // ordered by id
		if out[l.SupplementID] == nil {
			out[l.SupplementID] = map[time.Time]SupplementLogRow{}
		}
		out[l.SupplementID][l.Date] = l
	}
	return out
}

// ---------- /api/supplements ----------

func BuildSupplements(data Data, sd SupplementData, today time.Time) SupplementsResponse {
	from := today.Add(-regularityDays * day)
	resp := SupplementsResponse{
		Date:       iso(today),
		DayType:    dayTypeOf(data, today, today),
		Checklist:  []SupplementItem{},
		Regularity: RegularityRange{From: iso(from), To: iso(today.Add(-day)), Items: []SupplementRegularity{}},
		Stock:      []SupplementStock{},
		Proposal:   reminderProposal(data),
	}
	logs := latestLogs(sd.Logs)
	lasts := map[string]float64{} // supplement id -> days the stock lasts
	dayTypes := map[time.Time]*string{}
	for _, d := range metrics.Days(from, today.Add(-day)) {
		dayTypes[d] = dayTypeOf(data, d, today)
	}

	for _, s := range sd.Supplements {
		if !s.activeOn(today) {
			continue
		}
		if item, ok := checklistItem(s, logs[s.ID], resp.DayType, today); ok {
			resp.Checklist = append(resp.Checklist, item)
		}
		reg := regularity(s, logs[s.ID], dayTypes, from, today)
		resp.Regularity.Items = append(resp.Regularity.Items, reg)
		if st, days, ok := stockOf(s, sd.Stock, logs[s.ID], reg); ok {
			resp.Stock = append(resp.Stock, st)
			lasts[s.ID] = days
		}
	}

	sort.SliceStable(resp.Checklist, func(i, j int) bool {
		return slotOrder[resp.Checklist[i].Slot] < slotOrder[resp.Checklist[j].Slot]
	})
	sort.SliceStable(resp.Regularity.Items, func(i, j int) bool {
		return resp.Regularity.Items[i].Pct > resp.Regularity.Items[j].Pct
	})
	sort.SliceStable(resp.Stock, func(i, j int) bool {
		return lasts[resp.Stock[i].SupplementID] < lasts[resp.Stock[j].SupplementID]
	})
	return resp
}

// checklistItem is the supplement's entry for today: due by its rules, or already logged today anyway.
func checklistItem(s SupplementRow, logs map[time.Time]SupplementLogRow, dayType *string, today time.Time) (SupplementItem, bool) {
	rule := s.dueRule(dayType)
	l, logged := logs[today]
	if rule == nil && !logged {
		return SupplementItem{}, false
	}
	if rule == nil && len(s.Rules) > 0 {
		rule = &s.Rules[0]
	}
	item := SupplementItem{ID: s.ID, Name: s.Name, Dose: Dose{Amount: s.DoseAmount, Unit: s.DoseUnit},
		CountedInFitatu: s.CountedInFitatu, Status: "pending", Slot: "morning"}
	if rule != nil {
		item.Slot, item.RemindAt, item.Note = rule.Slot, rule.RemindAt, rule.Note
	}
	if logged {
		item.Status = map[string]string{SupplementTaken: "taken", SupplementSkipped: "skipped"}[l.Type]
		item.LoggedAt = ptr(l.CreatedAt.In(projector.Warsaw).Format(time.RFC3339))
	}
	return item, true
}

func regularity(s SupplementRow, logs map[time.Time]SupplementLogRow, dayTypes map[time.Time]*string,
	from, today time.Time) SupplementRegularity {
	when := "daily"
	if len(s.Rules) > 0 {
		when = s.Rules[0].When
	}
	r := SupplementRegularity{SupplementID: s.ID, Name: s.Name, When: when, Tone: "neutral"}
	for _, d := range metrics.Days(from, today.Add(-day)) {
		l, logged := logs[d]
		taken := logged && l.Type == SupplementTaken
		switch {
		case taken: // taken counts even on a day the rule did not ask for it
			r.Days = append(r.Days, "taken")
			r.Taken++
			r.Applicable++
		case s.activeOn(d) && s.dueRule(dayTypes[d]) != nil:
			r.Days = append(r.Days, "missed")
			r.Applicable++
		default:
			r.Days = append(r.Days, "na")
		}
	}
	if r.Applicable > 0 {
		r.Pct = int(math.Round(float64(r.Taken) / float64(r.Applicable) * 100))
		if r.Pct < regularityWarnPct {
			r.Tone = "warn"
		}
	}
	return r
}

// stockOf is the stock left (latest count minus doses taken after it) and how long it lasts; ok is false
// when no stock was ever entered.
func stockOf(s SupplementRow, stock []StockRow, logs map[time.Time]SupplementLogRow, reg SupplementRegularity) (SupplementStock, float64, bool) {
	var latest *StockRow
	for i := range stock {
		if stock[i].SupplementID == s.ID {
			latest = &stock[i] // ordered by created_at
		}
	}
	if latest == nil || s.PackPerDose <= 0 {
		return SupplementStock{}, 0, false
	}
	left := latest.Amount
	for _, l := range logs {
		if l.Type == SupplementTaken && l.CreatedAt.After(latest.CreatedAt) {
			left -= s.PackPerDose
		}
	}
	left = math.Max(0, left)
	st := SupplementStock{SupplementID: s.ID, Name: s.Name, Left: Dose{Amount: metrics.Round(left, 1), Unit: s.PackUnit}}
	st.Lasts.Count = int(math.Floor(left / s.PackPerDose))
	st.Lasts.Per = reg.When
	days := usesToDays(st.Lasts.Count, reg)
	st.Reorder = days < reorderDays
	return st, days, true
}

// usesToDays converts remaining uses into days: one per day for daily rules, otherwise at the rate the rule
// applied over the last 30 days.
func usesToDays(uses int, reg SupplementRegularity) float64 {
	if reg.When == "daily" {
		return float64(uses)
	}
	if reg.Applicable == 0 {
		return math.Inf(1)
	}
	return float64(uses) / (float64(reg.Applicable) / regularityDays)
}

// SupplementsSummary is the counter for the Jedzenie | Suplementy switch and the Today screen; nil when
// no supplements are set up.
func SupplementsSummary(s SupplementsResponse) *SupplementsToday {
	if len(s.Checklist) == 0 {
		return nil
	}
	out := &SupplementsToday{Total: len(s.Checklist)}
	for _, it := range s.Checklist {
		switch {
		case it.Status == "taken":
			out.Taken++
		case it.Status == "pending" && out.Next == nil:
			out.Next = ptr(it.Name)
		}
	}
	return out
}
