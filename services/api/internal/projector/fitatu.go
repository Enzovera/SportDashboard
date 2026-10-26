package projector

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// NutritionDay is one nutrition_day row (Fitatu diet_day, targets from the settings in force that day).
type NutritionDay struct {
	Date     time.Time
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
	Target   *DietTarget
	Meals    []Meal
}

// Meal is one nutrition_meal row: a Fitatu meal slot with at least one product.
type Meal struct {
	Slot     string
	SlotNo   int
	At       *time.Time
	Title    string
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
}

// DietTarget is the daily goal set in Fitatu (userDietSettings).
type DietTarget struct {
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
}

// Fitatu meal slots in the app's order (weekMealSchema numbers 1..6).
var mealSlots = map[string]int{"breakfast": 1, "second_breakfast": 2, "lunch": 3, "dinner": 4, "snack": 5, "supper": 6}

type dietDayBody struct {
	DietPlan map[string]struct {
		MealTime string `json:"mealTime"` // "HH:MM" or ""
		Items    []struct {
			Name         string   `json:"name"`
			Visible      *bool    `json:"visible"`
			Energy       *float64 `json:"energy"`
			Protein      *float64 `json:"protein"`
			Carbohydrate *float64 `json:"carbohydrate"`
			Fat          *float64 `json:"fat"`
		} `json:"items"`
	} `json:"dietPlan"`
}

// ParseDietDay returns the day's totals and meals; ok is false when nothing was logged that day.
func ParseDietDay(date time.Time, raw []byte) (NutritionDay, bool, error) {
	var b dietDayBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return NutritionDay{}, false, fmt.Errorf("diet_day: %w", err)
	}
	d := NutritionDay{Date: date}
	for slot, m := range b.DietPlan {
		no, known := mealSlots[slot]
		if !known {
			continue
		}
		meal := Meal{Slot: slot, SlotNo: no}
		var names []string
		for _, it := range m.Items {
			if it.Visible != nil && !*it.Visible {
				continue
			}
			names = append(names, strings.TrimSpace(it.Name))
			meal.Kcal += val(it.Energy)
			meal.ProteinG += val(it.Protein)
			meal.CarbsG += val(it.Carbohydrate)
			meal.FatG += val(it.Fat)
		}
		if len(names) == 0 {
			continue
		}
		meal.Title = strings.Join(names, ", ")
		if t, err := time.ParseInLocation("15:04", m.MealTime, Warsaw); err == nil {
			at := time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, Warsaw)
			meal.At = &at
		}
		d.Kcal += meal.Kcal
		d.ProteinG += meal.ProteinG
		d.CarbsG += meal.CarbsG
		d.FatG += meal.FatG
		meal.Kcal, meal.ProteinG, meal.CarbsG, meal.FatG = math.Round(meal.Kcal), round1(meal.ProteinG), round1(meal.CarbsG), round1(meal.FatG)
		d.Meals = append(d.Meals, meal)
	}
	if len(d.Meals) == 0 {
		return d, false, nil
	}
	sort.Slice(d.Meals, func(i, j int) bool { return d.Meals[i].SlotNo < d.Meals[j].SlotNo })
	d.Kcal, d.ProteinG, d.CarbsG, d.FatG = math.Round(d.Kcal), round1(d.ProteinG), round1(d.CarbsG), round1(d.FatG)
	return d, true, nil
}

type settingsBody struct {
	UserDietSettings struct {
		Energy             *float64 `json:"energy"`
		ProteinWeight      *float64 `json:"proteinWeight"`
		CarbohydrateWeight *float64 `json:"carbohydrateWeight"`
		FatWeight          *float64 `json:"fatWeight"`
	} `json:"userDietSettings"`
}

// ParseSettings returns the daily goal; nil when Fitatu has no energy goal set.
func ParseSettings(raw []byte) (*DietTarget, error) {
	var b settingsBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	s := b.UserDietSettings
	if s.Energy == nil || *s.Energy <= 0 {
		return nil, nil
	}
	return &DietTarget{Kcal: *s.Energy, ProteinG: val(s.ProteinWeight), CarbsG: val(s.CarbohydrateWeight), FatG: val(s.FatWeight)}, nil
}

// targetFor picks the settings in force on date: the latest on or before it, else the earliest known
// (ingest fetches settings only for recent days, older diet days still get a target).
func targetFor(settings map[time.Time]*DietTarget, date time.Time) *DietTarget {
	var best, earliest time.Time
	for d := range settings {
		if !d.After(date) && d.After(best) {
			best = d
		}
		if earliest.IsZero() || d.Before(earliest) {
			earliest = d
		}
	}
	if !best.IsZero() {
		return settings[best]
	}
	if !earliest.IsZero() {
		return settings[earliest]
	}
	return nil
}

func val(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
