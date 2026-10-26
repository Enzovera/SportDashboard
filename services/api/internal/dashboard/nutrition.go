package dashboard

import (
	"math"
	"time"

	"forma/api/internal/metrics"
	"forma/api/internal/projector"
)

const (
	nutritionRangeDays = 14
	// Hard protein floor of the validators (docs/design.md, "Walidatory"), g per kg of body weight.
	ProteinMinGPerKg = 1.6
	// A run at least this long makes a long-run day (the same threshold as the supplement rule "bieg > 90 min").
	longRunS = 90 * 60
)

// Day types: the nutrition target of a day.
const (
	DayRest        = "rest"
	DayStrength    = "strength"
	DayRun         = "run"
	DayStrengthRun = "strength_run"
	DayLongRun     = "long_run"
)

// inferDayType reads the day type from what was done, until the plan (stage 4) says it.
// A past day without activity is a rest day; today without activity yet is unknown (nil).
func inferDayType(data Data, date, today time.Time) *string {
	var run, long, strength bool
	for _, a := range data.Activities {
		if a.LocalDate.Equal(date) && isRun(a.Sport) {
			run = true
			if a.DurationS != nil && *a.DurationS >= longRunS {
				long = true
			}
		}
	}
	for _, w := range data.Workouts {
		if w.LocalDate.Equal(date) {
			strength = true
		}
	}
	switch {
	case long:
		return ptr(DayLongRun)
	case run && strength:
		return ptr(DayStrengthRun)
	case run:
		return ptr(DayRun)
	case strength:
		return ptr(DayStrength)
	case date.Before(today):
		return ptr(DayRest)
	default:
		return nil
	}
}

// weightAt is the body weight for g/kg on a date: the 7-day mean, else the last weigh-in before it, else the
// first one after it.
func weightAt(days []DayRow, date time.Time) *float64 {
	if t := buildWeight(days, date); t != nil {
		return ptr(t.Avg7dKg)
	}
	var last, first *float64
	for _, d := range days {
		if d.WeightKg == nil {
			continue
		}
		if !d.Date.After(date) {
			last = d.WeightKg
		} else if first == nil {
			first = d.WeightKg
		}
	}
	if last != nil {
		return last
	}
	return first
}

// ---------- /api/nutrition ----------

func BuildNutrition(data Data, today, now time.Time) NutritionResponse {
	from := today.Add(-(nutritionRangeDays - 1) * day)
	resp := NutritionResponse{
		Range:           "14d",
		From:            iso(from),
		To:              iso(today),
		Sync:            syncOf(BuildSync(data.Sync, now), "fitatu"),
		Days:            []NutritionDay{},
		Targets:         []DayTarget{},
		TargetProposals: []TargetProposal{},
		Meals:           []Meal{},
	}
	if len(data.Nutrition) == 0 {
		return resp // Fitatu not connected yet
	}

	byDate := map[time.Time]NutritionDayRow{}
	for _, n := range data.Nutrition {
		byDate[n.Date] = n
	}
	// Target in force on a date: the latest one known on or before it (days without a logged meal included).
	targetAt := func(date time.Time) *DietTarget {
		var t *DietTarget
		for _, n := range data.Nutrition {
			if !n.Date.After(date) && n.Target != nil {
				t = n.Target
			}
		}
		return t
	}

	var (
		n, gkgN, atMin                      int
		kcalSum, targetSum, gkgSum, gkgTarg float64
		longRun                             *LongRunCarbs
	)
	for _, date := range metrics.Days(from, today) {
		complete := date.Before(today)
		row, logged := byDate[date]
		target := row.Target
		if target == nil {
			target = targetAt(date)
		}
		nd := NutritionDay{Date: iso(date), DayType: inferDayType(data, date, today), Complete: complete,
			KcalTone: "neutral", ProteinTone: "neutral"}
		if target != nil {
			nd.KcalTarget = ptr(target.Kcal)
		}
		weight := weightAt(data.Days, date)
		if logged {
			nd.Kcal = ptr(row.Kcal)
			if weight != nil {
				nd.ProteinGPerKg = ptr(metrics.Round(row.ProteinG / *weight, 2))
				if *nd.ProteinGPerKg < ProteinMinGPerKg {
					nd.ProteinTone = "warn"
				}
			}
		}
		if complete && logged && target != nil && target.Kcal > 0 {
			pct := math.Round((row.Kcal - target.Kcal) / target.Kcal * 100)
			nd.KcalVsTargetPct = ptr(pct)
			longDay := nd.DayType != nil && *nd.DayType == DayLongRun
			if pct < 0 && longDay {
				nd.KcalTone = "warn" // validator: no deficit on a long-run day
			}

			n++
			kcalSum += row.Kcal
			targetSum += target.Kcal
			if weight != nil {
				gkgN++
				gkgSum += row.ProteinG / *weight
				gkgTarg += target.ProteinG / *weight
				if row.ProteinG / *weight >= ProteinMinGPerKg {
					atMin++
				}
				if longDay && target.CarbsG > 0 {
					longRun = &LongRunCarbs{Date: iso(date), GPerKg: round1(row.CarbsG / *weight),
						TargetGPerKg: round1(target.CarbsG / *weight), Tone: "neutral"}
					if longRun.GPerKg < longRun.TargetGPerKg {
						longRun.Tone = "warn"
					}
				}
			}
		}
		resp.Days = append(resp.Days, nd)
	}

	if n > 0 && gkgN > 0 {
		resp.Balance = &NutritionBalance{
			From:                iso(from),
			To:                  iso(today.Add(-day)),
			KcalAvg:             math.Round(kcalSum / float64(n)),
			KcalTargetAvg:       math.Round(targetSum / float64(n)),
			KcalVsTargetPct:     math.Round((kcalSum - targetSum) / targetSum * 100),
			ProteinAvgGPerKg:    metrics.Round(gkgSum/float64(gkgN), 2),
			ProteinTargetGPerKg: metrics.Round(gkgTarg/float64(gkgN), 2),
			ProteinMinGPerKg:    ProteinMinGPerKg,
			ProteinDaysAtMin:    atMin,
			Days:                n,
			LongRunCarbs:        longRun,
		}
	}

	if t := targetAt(today); t != nil {
		row := byDate[today]
		dt := DayRest
		if v := inferDayType(data, today, today); v != nil {
			dt = *v
		}
		resp.Today = &NutritionTodayDetail{
			DayType:  dt,
			Kcal:     NutritionRow{Eaten: row.Kcal, Target: t.Kcal},
			ProteinG: NutritionRow{Eaten: math.Round(row.ProteinG), Target: t.ProteinG},
			CarbsG:   NutritionRow{Eaten: math.Round(row.CarbsG), Target: t.CarbsG},
			FatG:     NutritionRow{Eaten: math.Round(row.FatG), Target: t.FatG},
		}
	}

	for _, m := range data.Meals {
		if !m.Date.Equal(today) {
			continue
		}
		meal := Meal{ID: iso(m.Date) + "-" + m.Slot, Slot: m.Slot, Title: m.Title, Kcal: m.Kcal,
			ProteinG: math.Round(m.ProteinG), CarbsG: math.Round(m.CarbsG), FatG: math.Round(m.FatG)}
		if m.At != nil {
			meal.At = ptr(m.At.In(projector.Warsaw).Format(time.RFC3339))
		}
		resp.Meals = append(resp.Meals, meal)
	}
	return resp
}
