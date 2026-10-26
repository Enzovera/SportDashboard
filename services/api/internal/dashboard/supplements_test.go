package dashboard

import (
	"testing"
	"time"
)

func supplementData() (Data, SupplementData) {
	longRun := run("lr", "2026-10-18", 21, 250)
	longRun.DurationS = fp(2 * 3600)
	data := Data{Activities: []ActivityRow{longRun, run("r1", "2026-10-25", 8, 90)}}

	start := d("2026-09-01")
	sd := SupplementData{
		Supplements: []SupplementRow{
			{ID: "magnesium", Name: "Magnez", DoseAmount: 300, DoseUnit: "mg", PackUnit: "tablet", PackPerDose: 1, ActiveFrom: start,
				Rules: []SupplementRuleRow{{When: "daily", Slot: "evening", RemindAt: sp("20:00"), Note: sp("przed snem")}}},
			{ID: "d3k2", Name: "Witamina D3 + K2", DoseAmount: 2000, DoseUnit: "IU", PackUnit: "capsule", PackPerDose: 1, ActiveFrom: start,
				Rules: []SupplementRuleRow{{When: "daily", Slot: "morning", RemindAt: sp("07:30")}}},
			{ID: "electrolytes", Name: "Elektrolity", DoseAmount: 1, DoseUnit: "sachet", PackUnit: "sachet", PackPerDose: 1, ActiveFrom: start,
				Rules: []SupplementRuleRow{{When: "long_run", Slot: "during"}}},
			{ID: "creatine", Name: "Kreatyna", DoseAmount: 5, DoseUnit: "g", PackUnit: "g", PackPerDose: 5, ActiveFrom: start,
				Rules: []SupplementRuleRow{{When: "strength_day", Slot: "after"}}},
			{ID: "old", Name: "Stary", DoseAmount: 1, DoseUnit: "tablet", PackUnit: "tablet", PackPerDose: 1, ActiveFrom: start,
				ActiveTo: tp(d("2026-10-01")), Rules: []SupplementRuleRow{{When: "daily", Slot: "morning"}}},
		},
		Stock: []StockRow{
			{SupplementID: "d3k2", Amount: 20, CreatedAt: d("2026-10-19")},
			{SupplementID: "electrolytes", Amount: 6, CreatedAt: d("2026-10-01")},
		},
	}
	log := func(id, date, typ string, h int) {
		sd.Logs = append(sd.Logs, SupplementLogRow{SupplementID: id, Date: d(date), Type: typ, CreatedAt: d(date).Add(time.Duration(h) * time.Hour)})
	}
	// D3: taken every day of the last 30 except one; today taken.
	for i := 1; i <= 30; i++ {
		if i != 10 {
			log("d3k2", iso(today.Add(-time.Duration(i)*day)), SupplementTaken, 7)
		}
	}
	log("d3k2", "2026-10-25", SupplementTaken, 7)
	// Magnesium: taken 20 of 30 days; ticked today, then unticked (the latest event wins).
	for i := 1; i <= 20; i++ {
		log("magnesium", iso(today.Add(-time.Duration(i)*day)), SupplementTaken, 20)
	}
	log("magnesium", "2026-10-25", SupplementTaken, 20)
	log("magnesium", "2026-10-25", SupplementSkipped, 21)
	log("electrolytes", "2026-10-18", SupplementTaken, 9)
	return data, sd
}

func TestBuildSupplements(t *testing.T) {
	data, sd := supplementData()
	s := BuildSupplements(data, sd, today)

	if s.DayType == nil || *s.DayType != DayRun {
		t.Fatalf("day type %v", s.DayType)
	}
	// Run day: no electrolytes (long run only), no creatine (strength day), no inactive supplement.
	if len(s.Checklist) != 2 || s.Checklist[0].ID != "d3k2" || s.Checklist[1].ID != "magnesium" {
		t.Fatalf("checklist: %+v", s.Checklist)
	}
	d3, mg := s.Checklist[0], s.Checklist[1]
	if d3.Status != "taken" || *d3.LoggedAt != "2026-10-25T08:00:00+01:00" || *d3.RemindAt != "07:30" || d3.Dose != (Dose{2000, "IU"}) {
		t.Errorf("d3: %+v", d3)
	}
	if mg.Status != "skipped" || *mg.Note != "przed snem" {
		t.Errorf("magnesium: %+v", mg)
	}

	if s.Regularity.From != "2026-09-25" || s.Regularity.To != "2026-10-24" {
		t.Errorf("regularity range %s..%s", s.Regularity.From, s.Regularity.To)
	}
	reg := map[string]SupplementRegularity{}
	for _, r := range s.Regularity.Items {
		reg[r.SupplementID] = r
		if len(r.Days) != 30 {
			t.Errorf("%s: %d days", r.SupplementID, len(r.Days))
		}
	}
	if r := reg["d3k2"]; r.Taken != 29 || r.Applicable != 30 || r.Pct != 97 || r.Tone != "neutral" {
		t.Errorf("d3 regularity: %+v", r)
	}
	if r := reg["magnesium"]; r.Pct != 67 || r.Tone != "warn" {
		t.Errorf("magnesium regularity: %+v", r)
	}
	if r := reg["electrolytes"]; r.Taken != 1 || r.Applicable != 1 || r.Days[23] != "taken" || r.Days[0] != "na" {
		t.Errorf("electrolytes regularity: %+v", r)
	}
	if r := reg["creatine"]; r.Applicable != 0 || r.Pct != 0 || r.Tone != "neutral" {
		t.Errorf("creatine regularity: %+v", r)
	}
	if _, ok := reg["old"]; ok {
		t.Error("inactive supplement in regularity")
	}

	// D3: 20 bought at midnight 19.10, taken 19.10..25.10 after it = 7 → 13 left, under 14 days: reorder.
	// Electrolytes: 6 − 1 = 5 long runs at 1 per 30 days → far off. Most urgent first.
	if len(s.Stock) != 2 || s.Stock[0].SupplementID != "d3k2" {
		t.Fatalf("stock: %+v", s.Stock)
	}
	if st := s.Stock[0]; st.Left != (Dose{13, "capsule"}) || st.Lasts.Count != 13 || st.Lasts.Per != "daily" || !st.Reorder {
		t.Errorf("d3 stock: %+v", st)
	}
	if st := s.Stock[1]; st.Lasts.Count != 5 || st.Lasts.Per != "long_run" || st.Reorder {
		t.Errorf("electrolytes stock: %+v", st)
	}

	sum := SupplementsSummary(s)
	if sum == nil || sum.Taken != 1 || sum.Total != 2 || sum.Next != nil {
		t.Errorf("summary: %+v", sum)
	}
}

func TestRuleApplies(t *testing.T) {
	tests := []struct {
		when    string
		dayType *string
		want    bool
	}{
		{"daily", nil, true},
		{"run_day", nil, false},
		{"long_run", ptr(DayLongRun), true},
		{"long_run", ptr(DayRun), false},
		{"quality_run", ptr(DayStrengthRun), true},
		{"strength_day", ptr(DayStrengthRun), true},
		{"strength_day", ptr(DayRun), false},
		{"training_day", ptr(DayStrength), true},
		{"training_day", ptr(DayRest), false},
	}
	for _, tt := range tests {
		if got := ruleApplies(tt.when, tt.dayType); got != tt.want {
			t.Errorf("%s on %v = %v", tt.when, tt.dayType, got)
		}
	}
}
