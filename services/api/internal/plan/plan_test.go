package plan

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

func fp(v float64) *float64 { return &v }
func sp(v string) *string   { return &v }

func runDay(date string, km float64, dayType string) Day {
	return Day{Date: Date{d(date)}, DayType: dayType, Sessions: []Session{{Kind: "run", Title: "bieg", DistanceKm: fp(km)}}}
}

// week of 46 km starting on Monday `from`: Mo 6, Tu 10, We strength, Th rest, Fr 6, Sa 8, Su long 16.
func week(from string) []Day {
	m := d(from)
	at := func(i int) string { return m.AddDate(0, 0, i).Format(time.DateOnly) }
	return []Day{
		runDay(at(0), 6, DayStrengthRun), runDay(at(1), 10, DayRun),
		{Date: Date{d(at(2))}, DayType: DayStrength, Sessions: []Session{{Kind: "strength", Title: "Legs", Detail: sp("16 serii roboczych")}}},
		runDay(at(4), 6, DayStrengthRun), runDay(at(5), 8, DayRun), runDay(at(6), 16, DayLongRun),
	}
}

func event(seq int, typ string, payload any, at time.Time) Event {
	b, _ := json.Marshal(payload)
	return Event{StreamID: "s1", Seq: seq, Type: typ, Payload: b, CreatedAt: at}
}

func created() Created {
	return Created{
		V:      1,
		Goal:   Goal{Race: "Maraton", Date: Date{d("2027-03-14")}},
		Blocks: []Block{{Kind: "base", From: Date{d("2026-10-26")}, To: Date{d("2026-12-06")}}},
		Targets: []Target{
			{DayType: DayLongRun, Kcal: fp(2950), ProteinG: fp(145), CarbsG: fp(412), FatG: fp(80)},
			{DayType: DayRest, Kcal: fp(2300), ProteinG: fp(145), CarbsG: fp(255), FatG: fp(80)},
		},
		Days: append(week("2026-10-19"), week("2026-10-26")...),
	}
}

var t0 = time.Date(2026, 10, 18, 19, 0, 0, 0, time.UTC)

func TestFold(t *testing.T) {
	carbs := Proposed{V: 1, ProposalID: "p-carbs", Kind: "nutrition", EffectiveFrom: &Date{d("2026-11-01")},
		Change: Change{Targets: []Target{{DayType: DayLongRun, CarbsG: fp(460), Kcal: fp(3140)}}}}
	legs := Proposed{V: 1, ProposalID: "p-legs", Kind: "strength",
		Change: Change{Days: []Day{{Date: Date{d("2026-10-28")}, DayType: DayStrength,
			Sessions: []Session{{Kind: "strength", Title: "Legs", Detail: sp("12 serii roboczych")}}}}}}
	st, err := Fold([]Event{
		event(1, PlanCreated, created(), t0),
		event(2, PlanChangeProposed, carbs, t0.Add(time.Hour)),
		event(3, PlanChangeProposed, legs, t0.Add(2*time.Hour)),
		event(4, PlanChangeAccepted, Decided{V: 1, ProposalID: "p-carbs"}, t0.Add(3*time.Hour)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.LastSeq != 4 || len(st.Pending()) != 1 || st.Pending()[0].ProposalID != "p-legs" {
		t.Fatalf("state: seq %d pending %+v", st.LastSeq, st.Pending())
	}
	// Accepted target applies from its effective date only, and keeps the fields it does not change.
	if tg := st.TargetAt(d("2026-10-25"), DayLongRun); tg.CarbsG != 412 {
		t.Errorf("before: %+v", tg)
	}
	if tg := st.TargetAt(d("2026-11-01"), DayLongRun); tg.CarbsG != 460 || tg.Kcal != 3140 || tg.ProteinG != 145 {
		t.Errorf("after: %+v", tg)
	}
	if st.Day(d("2026-10-29")).DayType != DayRest || st.Day(d("2026-10-28")).Sessions[0].Detail == nil {
		t.Error("days")
	}
	if b, n := st.BlockAt(d("2026-11-03")); b == nil || b.Kind != "base" || n != 2 {
		t.Errorf("block: %v %d", b, n)
	}

	_, err = Fold([]Event{event(1, PlanCreated, created(), t0), event(2, PlanChangeAccepted, Decided{ProposalID: "x"}, t0)})
	if err == nil {
		t.Error("decision of an unknown proposal must fail")
	}
}

func TestValidate(t *testing.T) {
	st, _ := Fold([]Event{event(1, PlanCreated, created(), t0)})
	lim := Limits{KcalFloor: 1800, MaintenanceKcal: fp(2600), WeightKg: fp(76)}
	from := d("2026-10-26")
	long := func(km float64) Change { return Change{Days: []Day{runDay("2026-11-01", km, DayLongRun)}} }
	allDays := func() Change {
		c := Change{}
		for i := 0; i < 7; i++ {
			c.Days = append(c.Days, runDay(d("2026-10-26").AddDate(0, 0, i).Format(time.DateOnly), 6, DayRun))
		}
		return c
	}

	tests := []struct {
		name   string
		change Change
		want   map[string]string // rule -> status
		text   string            // substring of some check text
	}{
		{"long run 16 → 22: +13% blocked, long share warn", long(22),
			map[string]string{RuleKmIncrease: Fail, RuleLongRunShare: Warn, RuleRestDay: Pass}, "przyrost 46 → 52 km (+13%, limit 10%)"},
		{"long run 16 → 18: +4% passes, long share warns", long(18),
			map[string]string{RuleKmIncrease: Pass, RuleLongRunShare: Warn}, "długie = 38% tygodnia (limit 35%)"},
		{"no rest day", allDays(), map[string]string{RuleRestDay: Fail}, "brak dnia wolnego w tygodniu 26.10–1.11"},
		{"carbs up on long run", Change{Targets: []Target{{DayType: DayLongRun, CarbsG: fp(460), Kcal: fp(3140)}}},
			map[string]string{RuleKcalFloor: Pass, RuleLongRunDeficit: Pass, RuleProteinMin: Pass}, "brak deficytu"},
		{"long run deficit", Change{Targets: []Target{{DayType: DayLongRun, Kcal: fp(2400)}}},
			map[string]string{RuleLongRunDeficit: Fail}, "2400 < 2600"},
		{"kcal floor", Change{Targets: []Target{{DayType: DayRest, Kcal: fp(1500)}}},
			map[string]string{RuleKcalFloor: Fail}, "1500 kcal poniżej progu 1800 kcal"},
		{"protein below 1,6 g/kg warns", Change{Targets: []Target{{DayType: DayRest, ProteinG: fp(110)}}},
			map[string]string{RuleProteinMin: Warn, RuleKcalFloor: Pass}, "1,45 g/kg"},
		{"reminder moved", Change{Reminder: &Reminder{SupplementID: "magnesium", From: "20:00", To: "21:30"}},
			map[string]string{RuleReminderTime: Pass}, "20:00 → 21:30"},
		{"reminder bad time", Change{Reminder: &Reminder{SupplementID: "magnesium", From: "20:00", To: "25:00"}},
			map[string]string{RuleReminderTime: Fail}, "HH:MM"},
		{"empty change", Change{}, map[string]string{RuleChange: Fail}, "nic nie zmienia"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checks := Validate(st, tt.change, from, lim)
			got := map[string]string{}
			var texts []string
			for _, c := range checks {
				if prev, ok := got[c.Rule]; !ok || prev != Fail {
					got[c.Rule] = c.Status
				}
				texts = append(texts, c.Text)
			}
			for rule, want := range tt.want {
				if got[rule] != want {
					t.Errorf("%s = %q, want %q (%+v)", rule, got[rule], want, checks)
				}
			}
			if !strings.Contains(strings.Join(texts, " | "), tt.text) {
				t.Errorf("no %q in %v", tt.text, texts)
			}
		})
	}
}

func TestValidateDeloadAndTaper(t *testing.T) {
	c := created()
	// Week of 19.10 is a deload (30 km after 46): the week after is compared with 46.
	c.Days = append(week("2026-10-12"), runDay("2026-10-19", 30, DayRun))
	c.Days = append(c.Days, week("2026-10-26")...)
	st, _ := Fold([]Event{event(1, PlanCreated, c, t0)})
	checks := Validate(st, Change{Days: []Day{runDay("2026-11-01", 20, DayLongRun)}}, d("2026-10-26"), Limits{})
	if checks[0].Rule != RuleKmIncrease || checks[0].Status != Pass {
		t.Errorf("deload: %+v", checks)
	}

	// Taper: the last 3 weeks before 14.03.2027 must not grow.
	c = created()
	c.Days = append(week("2027-02-22"), week("2027-03-01")...)
	st, _ = Fold([]Event{event(1, PlanCreated, c, t0)})
	checks = Validate(st, Change{Days: []Day{runDay("2027-03-07", 18, DayLongRun)}}, d("2027-03-01"), Limits{})
	var taper *Check
	for i := range checks {
		if checks[i].Rule == RuleTaper {
			taper = &checks[i]
		}
	}
	if taper == nil || taper.Status != Fail {
		t.Errorf("taper: %+v", checks)
	}
}
