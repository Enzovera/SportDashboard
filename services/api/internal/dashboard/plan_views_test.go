package dashboard

import (
	"encoding/json"
	"testing"
	"time"

	"forma/api/internal/plan"
)

func pfp(v float64) *float64 { return &v }

func pday(date, dayType string, sessions ...plan.Session) plan.Day {
	return plan.Day{Date: plan.Date{Time: d(date)}, DayType: dayType, Sessions: sessions}
}

func prun(title string, km float64) plan.Session {
	return plan.Session{Kind: "run", Title: title, DistanceKm: pfp(km)}
}

func pstrength(title string) plan.Session { return plan.Session{Kind: "strength", Title: title} }

func planEvent(seq int, typ string, payload any, at time.Time) plan.Event {
	b, _ := json.Marshal(payload)
	return plan.Event{StreamID: "s1", Seq: seq, Type: typ, Payload: b, CreatedAt: at}
}

// testPlan: today = Sunday 25.10 (long run), base block from 26.10, two pending proposals (as in the mock-ups).
func testPlan(t *testing.T) *plan.State {
	t.Helper()
	week := func(mon string) []plan.Day {
		m := d(mon)
		at := func(i int) string { return iso(m.Add(time.Duration(i) * day)) }
		return []plan.Day{
			pday(at(0), plan.DayStrengthRun, pstrength("Push"), prun("bieg łatwy", 6)),
			pday(at(1), plan.DayRun, prun("Interwały", 10)),
			pday(at(2), plan.DayStrength, pstrength("Legs")),
			pday(at(4), plan.DayStrengthRun, pstrength("Pull"), prun("bieg łatwy", 6)),
			pday(at(5), plan.DayRun, prun("Tempo", 8)),
			pday(at(6), plan.DayLongRun, prun("Długie wybieganie", 16)),
		}
	}
	targets := []plan.Target{
		{DayType: plan.DayRest, Kcal: pfp(2300), ProteinG: pfp(145), CarbsG: pfp(255), FatG: pfp(80)},
		{DayType: plan.DayStrength, Kcal: pfp(2600), ProteinG: pfp(145), CarbsG: pfp(330), FatG: pfp(80)},
		{DayType: plan.DayRun, Kcal: pfp(2700), ProteinG: pfp(145), CarbsG: pfp(355), FatG: pfp(80)},
		{DayType: plan.DayStrengthRun, Kcal: pfp(2750), ProteinG: pfp(145), CarbsG: pfp(370), FatG: pfp(80)},
		{DayType: plan.DayLongRun, Kcal: pfp(2950), ProteinG: pfp(145), CarbsG: pfp(412), FatG: pfp(80)},
	}
	created := plan.Created{V: 1,
		Goal:    plan.Goal{Race: "Maraton", Date: plan.Date{Time: d("2027-03-14")}},
		Blocks:  []plan.Block{{Kind: "base", From: plan.Date{Time: d("2026-10-26")}, To: plan.Date{Time: d("2026-12-06")}}},
		Targets: targets,
		Days:    append(week("2026-10-19"), week("2026-10-26")...),
	}
	legs := plan.Proposed{V: 1, ProposalID: "p-legs", Kind: "strength", Origin: "daily_run", Title: "Śr 28.10, Legs: 16 → 12 serii",
		Summary: "12 serii", Change: plan.Change{Days: []plan.Day{
			pday("2026-10-28", plan.DayStrength, plan.Session{Kind: "strength", Title: "Legs", Short: ptr("Legs 12 serii")})}}}
	carbs := plan.Proposed{V: 1, ProposalID: "p-carbs", Kind: "nutrition", Origin: "daily_run", Summary: "W 460 g, 3 140 kcal",
		EffectiveFrom: &plan.Date{Time: d("2026-11-01")},
		Change:        plan.Change{Targets: []plan.Target{{DayType: plan.DayLongRun, Kcal: pfp(3140), CarbsG: pfp(460)}}}}
	reminder := plan.Proposed{V: 1, ProposalID: "p-mg", Kind: "nutrition", Origin: "weekly_run", Rationale: "Magnez pomijasz.",
		Change: plan.Change{Reminder: &plan.Reminder{SupplementID: "magnesium", SupplementName: "Magnez", From: "20:00", To: "21:30"}}}
	old := plan.Proposed{V: 1, ProposalID: "p-old", Kind: "run", Origin: "daily_run", Title: "Dodatkowy bieg"}
	t0 := time.Date(2026, 10, 18, 19, 0, 0, 0, time.UTC)
	st, err := plan.Fold([]plan.Event{
		planEvent(1, plan.PlanCreated, created, t0),
		planEvent(2, plan.PlanChangeProposed, old, t0.Add(time.Hour)),
		planEvent(3, plan.PlanChangeRejected, plan.Decided{V: 1, ProposalID: "p-old", Comment: ptr("czwartek zostaje wolny")}, t0.Add(2*time.Hour)),
		planEvent(4, plan.PlanChangeProposed, legs, d("2026-10-24").Add(5*time.Hour)),
		planEvent(5, plan.PlanChangeProposed, carbs, d("2026-10-25").Add(5*time.Hour)),
		planEvent(6, plan.PlanChangeProposed, reminder, d("2026-10-25").Add(6*time.Hour)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func TestBuildPlan(t *testing.T) {
	data := sample()
	data.Plan = testPlan(t)
	p := BuildPlan(data, today)

	if p.Plan == nil || *p.ProposalsPending != 3 {
		t.Fatalf("plan: %+v", p)
	}
	g := p.Plan.Goal
	if g.WeeksLeft != 20 || g.Date != "2027-03-14" || g.Secondary == nil || len(p.Plan.Blocks) != 1 || p.Plan.Blocks[0].Weeks != 6 {
		t.Errorf("goal/blocks: %+v %+v", g, p.Plan.Blocks)
	}
	cur, next := p.Plan.CurrentWeek, p.Plan.NextWeek
	if cur == nil || cur.Block != nil || cur.PlannedKm != 46 || cur.PlannedStrength != 3 || *cur.DoneKm != 24.2 || *cur.DoneStrength != 2 {
		t.Fatalf("current week: %+v", cur)
	}
	if mon := cur.Days[0]; mon.Done == nil || *mon.Done.RunKm != 6 || *mon.Done.StrengthSets != 18 || mon.Target.Kcal != 2750 {
		t.Errorf("monday: %+v", mon)
	}
	if sun := cur.Days[6]; sun.Done != nil || *sun.RunKm != 16 || sun.DayType != plan.DayLongRun {
		t.Errorf("today: %+v", sun)
	}
	if next == nil || *next.Block != "base" || *next.WeekInBlock != 1 || next.DoneKm != nil {
		t.Fatalf("next week: %+v", next)
	}
	if wed := next.Days[2]; len(wed.Proposals) != 1 || wed.Proposals[0].Summary != "12 serii" {
		t.Errorf("wednesday proposals: %+v", wed.Proposals)
	}
	if sun := next.Days[6]; len(sun.Proposals) != 1 || sun.Proposals[0].ID != "p-carbs" || sun.Target.CarbsG != 412 {
		t.Errorf("sunday: %+v", sun)
	}
	if thu := next.Days[3]; thu.DayType != plan.DayRest || len(thu.Sessions) != 0 || thu.Target.Kcal != 2300 {
		t.Errorf("rest day: %+v", thu)
	}

	if none := BuildPlan(sample(), today); none.Plan != nil || none.ProposalsPending != nil {
		t.Errorf("without a plan: %+v", none)
	}
}

func TestBuildProposals(t *testing.T) {
	st := testPlan(t)
	blocked := []plan.BlockedProposal{
		{ID: "b1", Title: "Nd 1.11: długie 16 → 22 km", ProposedAt: now.Add(-time.Hour), Checks: []plan.Check{{Rule: plan.RuleKmIncrease, Status: plan.Fail}}},
		{ID: "b0", ProposedAt: now.Add(-10 * 24 * time.Hour)},
	}
	p := BuildProposals(Data{Plan: st}, blocked, today)

	if len(p.Pending) != 3 || p.Pending[0].ID != "p-mg" || p.Pending[2].ID != "p-legs" || *p.Pending[1].EffectiveFrom != "2026-11-01" {
		t.Fatalf("pending: %+v", p.Pending)
	}
	if len(p.Blocked) != 1 || p.Blocked[0].ID != "b1" {
		t.Errorf("blocked: %+v", p.Blocked)
	}
	if len(p.History) != 1 || p.History[0].Decision != "rejected" || *p.History[0].Comment != "czwartek zostaje wolny" {
		t.Errorf("history: %+v", p.History)
	}
	w := p.Week
	if w == nil || w.From != "2026-10-26" || *w.Block != "Baza · tydz. 1" || len(w.Days) != 7 {
		t.Fatalf("week: %+v", w)
	}
	if mon := w.Days[0]; mon.Now.Session != "Push + bieg łatwy 6 km" || len(mon.Changes) != 0 {
		t.Errorf("monday: %+v", mon)
	}
	if wed := w.Days[2]; wed.Now.Session != "Legs" || len(wed.Changes) != 1 || wed.Changes[0].After.Session != "Legs 12 serii" {
		t.Errorf("wednesday: %+v", wed)
	}
	if sun := w.Days[6]; len(sun.Changes) != 1 || sun.Changes[0].After.Kcal != 3140 || sun.Changes[0].After.CarbsG != 460 || *sun.Changes[0].After.RunKm != 16 {
		t.Errorf("sunday: %+v", sun)
	}
	if thu := w.Days[3]; thu.Now.Session != "Odpoczynek" || thu.Now.RunKm != nil {
		t.Errorf("rest: %+v", thu)
	}

	if none := BuildProposals(Data{}, nil, today); none.Week != nil || len(none.Pending) != 0 || none.History == nil {
		t.Errorf("without a plan: %+v", none)
	}
}

func TestPlanDrivesNutritionAndSupplements(t *testing.T) {
	data := nutritionData()
	data.Plan = testPlan(t)
	n := BuildNutrition(data, today, now)
	if n.Today == nil || n.Today.DayType != plan.DayLongRun || n.Today.Kcal.Target != 2950 {
		t.Errorf("today: %+v", n.Today)
	}
	if len(n.Targets) != 5 || n.Targets[4].DayType != plan.DayLongRun || len(n.TargetProposals) != 2 {
		t.Errorf("targets: %+v %+v", n.Targets, n.TargetProposals)
	}
	if d := n.Days[12]; *d.DayType != plan.DayRun || *d.KcalTarget != 2700 { // Sat 24.10: plan says Tempo
		t.Errorf("24.10: %+v", d)
	}

	_, sd := supplementData()
	s := BuildSupplements(data, sd, today)
	if s.Proposal == nil || s.Proposal.ID != "p-mg" || s.Proposal.ProposedTime != "21:30" {
		t.Errorf("proposal: %+v", s.Proposal)
	}
	if len(s.Checklist) != 3 || s.Checklist[1].ID != "electrolytes" { // long run day per plan
		t.Errorf("checklist: %+v", s.Checklist)
	}
}
