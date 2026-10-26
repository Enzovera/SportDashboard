package dashboard

import (
	"testing"
	"time"
)

func nutritionData() Data {
	target := &DietTarget{Kcal: 2700, ProteinG: 140, CarbsG: 350, FatG: 80}
	longRun := run("lr", "2026-10-18", 21, 250)
	longRun.DurationS = fp(2 * 3600)
	data := Data{
		Activities: []ActivityRow{longRun, run("r1", "2026-10-24", 8, 90)},
		Workouts:   []WorkoutRow{{ID: "w1", LocalDate: d("2026-10-24"), StartedAt: now}},
		Nutrition: []NutritionDayRow{
			{Date: d("2026-10-18"), Kcal: 2600, ProteinG: 126, CarbsG: 315, FatG: 70, Target: target},
			{Date: d("2026-10-22"), Kcal: 2700, ProteinG: 98, CarbsG: 300, FatG: 90, Target: target},
			{Date: d("2026-10-24"), Kcal: 2565, ProteinG: 140, CarbsG: 330, FatG: 75, Target: target},
			{Date: d("2026-10-25"), Kcal: 820, ProteinG: 38.4, CarbsG: 112, FatG: 24, Target: target},
		},
		Meals: []MealRow{
			{Date: d("2026-10-24"), Slot: "dinner", Title: "Wczoraj", Kcal: 900},
			{Date: d("2026-10-25"), Slot: "breakfast", At: tp(time.Date(2026, 10, 25, 6, 40, 0, 0, time.UTC)),
				Title: "Owsianka", Kcal: 820, ProteinG: 38.4, CarbsG: 112, FatG: 24},
		},
	}
	for _, s := range []string{"2026-10-17", "2026-10-19", "2026-10-21", "2026-10-23"} {
		data.Days = append(data.Days, DayRow{Date: d(s), WeightKg: fp(70)})
	}
	return data
}

func TestBuildNutrition(t *testing.T) {
	n := BuildNutrition(nutritionData(), today, now)

	if n.From != "2026-10-12" || n.To != "2026-10-25" || len(n.Days) != 14 {
		t.Fatalf("range %s..%s, %d days", n.From, n.To, len(n.Days))
	}
	byDate := map[string]NutritionDay{}
	for _, x := range n.Days {
		byDate[x.Date] = x
	}
	tests := []struct {
		date, dayType, kcalTone, proteinTone string
		pct                                  *float64
		gkg                                  *float64
	}{
		{"2026-10-18", DayLongRun, "warn", "neutral", fp(-4), fp(1.8)}, // any deficit on a long-run day
		{"2026-10-22", DayRest, "neutral", "warn", fp(0), fp(1.4)},     // below 1,6 g/kg
		{"2026-10-24", DayStrengthRun, "neutral", "neutral", fp(-5), fp(2)},
		{"2026-10-20", DayRest, "neutral", "neutral", nil, nil}, // nothing logged
	}
	for _, tt := range tests {
		x := byDate[tt.date]
		if *x.DayType != tt.dayType || x.KcalTone != tt.kcalTone || x.ProteinTone != tt.proteinTone ||
			!eqp(x.KcalVsTargetPct, tt.pct) || !eqp(x.ProteinGPerKg, tt.gkg) || *x.KcalTarget != 2700 {
			t.Errorf("%s: %+v", tt.date, x)
		}
	}
	last := n.Days[13]
	if last.Complete || last.DayType != nil || last.KcalVsTargetPct != nil || *last.Kcal != 820 {
		t.Errorf("today: %+v", last)
	}

	b := n.Balance
	if b == nil || b.Days != 3 || b.KcalAvg != 2622 || b.KcalTargetAvg != 2700 || b.KcalVsTargetPct != -3 ||
		b.ProteinAvgGPerKg != 1.73 || b.ProteinTargetGPerKg != 2 || b.ProteinDaysAtMin != 2 || b.To != "2026-10-24" {
		t.Fatalf("balance: %+v", b)
	}
	if lr := b.LongRunCarbs; lr == nil || lr.Date != "2026-10-18" || lr.GPerKg != 4.5 || lr.TargetGPerKg != 5 || lr.Tone != "warn" {
		t.Errorf("long run carbs: %+v", b.LongRunCarbs)
	}

	if n.Today == nil || n.Today.DayType != DayRest || n.Today.Kcal != (NutritionRow{820, 2700}) || n.Today.ProteinG != (NutritionRow{38, 140}) {
		t.Errorf("today detail: %+v", n.Today)
	}
	if len(n.Meals) != 1 || n.Meals[0].ID != "2026-10-25-breakfast" || *n.Meals[0].At != "2026-10-25T07:40:00+01:00" {
		t.Errorf("meals: %+v", n.Meals)
	}
}

func TestBuildNutritionWithoutFitatu(t *testing.T) {
	n := BuildNutrition(sample(), today, now)
	if n.Today != nil || n.Balance != nil || len(n.Days) != 0 || n.Meals == nil || n.Targets == nil || len(n.Sync) != 0 {
		t.Errorf("unexpected: %+v", n)
	}
}
