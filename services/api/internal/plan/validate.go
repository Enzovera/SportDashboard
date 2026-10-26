package plan

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Validator keys and starting values (docs/design.md, "Walidatory i bezpieczeństwo"). Hard rules live here,
// not in the prompt: a `fail` keeps the proposal out of the queue and the coach gets the text as the reason.
const (
	RuleKmIncrease     = "km_increase"
	RuleLongRunShare   = "long_run_share"
	RuleRestDay        = "rest_day"
	RuleTaper          = "taper"
	RuleKcalFloor      = "kcal_floor"
	RuleLongRunDeficit = "long_run_deficit"
	RuleProteinMin     = "protein_min"
	RuleReminderTime   = "reminder_time"
	RuleChange         = "change"

	MaxKmIncrease   = 0.10
	MaxLongRunShare = 0.35
	// A week below this share of the one before is a deload: the next week is compared with the week before it.
	DeloadShare   = 0.90
	TaperWeeks    = 3
	ProteinMinGKg = 1.6
)

const (
	Pass = "pass"
	Warn = "warn"
	Fail = "fail"
)

// Limits are the inputs of the validators that come from outside the plan.
type Limits struct {
	KcalFloor       float64  // configured, not below the estimated BMR
	MaintenanceKcal *float64 // estimated daily expenditure; nil = unknown
	WeightKg        *float64 // 7-day mean; nil = unknown
}

// Validate checks a change against the plan it would be applied to (from `from` on).
func Validate(s *State, c Change, from time.Time, lim Limits) []Check {
	if len(c.Days) == 0 && len(c.Targets) == 0 && c.Reminder == nil {
		return []Check{{RuleChange, Fail, "propozycja nic nie zmienia"}}
	}
	var checks []Check
	if len(c.Days) > 0 {
		checks = append(checks, validateDays(s, s.With(c, from), c)...)
	}
	if len(c.Targets) > 0 {
		checks = append(checks, validateTargets(s.With(c, from), c, from, lim)...)
	}
	if c.Reminder != nil {
		checks = append(checks, validateReminder(*c.Reminder))
	}
	return checks
}

// Blocked reports whether any check failed.
func Blocked(checks []Check) bool {
	for _, c := range checks {
		if c.Status == Fail {
			return true
		}
	}
	return false
}

func weekStart(d time.Time) time.Time {
	return d.AddDate(0, 0, -((int(d.Weekday()) + 6) % 7))
}

func weekKm(s *State, start time.Time) float64 {
	var km float64
	for i := 0; i < 7; i++ {
		km += s.Day(start.AddDate(0, 0, i)).RunKm()
	}
	return km
}

func validateDays(before, after *State, c Change) []Check {
	weeks := map[time.Time]bool{}
	var order []time.Time
	for _, d := range c.Days {
		w := weekStart(d.Date.Time)
		if !weeks[w] {
			weeks[w] = true
			order = append(order, w)
		}
	}
	var checks []Check
	for _, w := range order {
		km := weekKm(after, w)
		prev := weekKm(after, w.AddDate(0, 0, -7))
		base := prev
		if prev2 := weekKm(after, w.AddDate(0, 0, -14)); prev2 > 0 && prev < prev2*DeloadShare {
			base = prev2 // back from a deload week
		}
		switch old := weekKm(before, w); {
		case base > 0 && km > base*(1+MaxKmIncrease):
			checks = append(checks, Check{RuleKmIncrease, Fail, fmt.Sprintf("przyrost %s → %s km (+%d%%, limit %d%%)",
				num(base), num(km), int(math.Round((km/base-1)*100)), int(MaxKmIncrease*100))})
		case km == old:
			checks = append(checks, Check{RuleKmIncrease, Pass, fmt.Sprintf("kilometraż bez zmian (%s km)", num(km))})
		default:
			checks = append(checks, Check{RuleKmIncrease, Pass, fmt.Sprintf("kilometraż %s → %s km", num(old), num(km))})
		}

		var longest float64
		for i := 0; i < 7; i++ {
			for _, x := range after.Day(w.AddDate(0, 0, i)).Sessions {
				if x.Kind == "run" && x.DistanceKm != nil && *x.DistanceKm > longest {
					longest = *x.DistanceKm
				}
			}
		}
		if km > 0 {
			share := longest / km
			st := Pass
			text := fmt.Sprintf("długie = %d%% tygodnia", int(math.Round(share*100)))
			if share > MaxLongRunShare {
				st, text = Warn, text+fmt.Sprintf(" (limit %d%%)", int(MaxLongRunShare*100))
			}
			checks = append(checks, Check{RuleLongRunShare, st, text})
		}

		var rest []string
		for i := 0; i < 7; i++ {
			if d := w.AddDate(0, 0, i); len(after.Day(d).Sessions) == 0 {
				rest = append(rest, weekday[d.Weekday()])
			}
		}
		if len(rest) == 0 {
			checks = append(checks, Check{RuleRestDay, Fail, "brak dnia wolnego w tygodniu " + span(w)})
		} else {
			checks = append(checks, Check{RuleRestDay, Pass, "dzień wolny: " + strings.Join(rest, ", ")})
		}

		race := after.Goal.Date.Time
		if !race.IsZero() && !w.Before(weekStart(race).AddDate(0, 0, -7*(TaperWeeks-1))) && !w.After(race) {
			if prev > 0 && km > prev {
				checks = append(checks, Check{RuleTaper, Fail, fmt.Sprintf("taper: %s → %s km, objętość nie może rosnąć", num(prev), num(km))})
			} else {
				checks = append(checks, Check{RuleTaper, Pass, "taper: objętość nie rośnie"})
			}
		}
	}
	return checks
}

func validateTargets(after *State, c Change, from time.Time, lim Limits) []Check {
	var checks []Check
	floorOK, longRun := true, false
	var low []string
	for _, t := range c.Targets {
		nt := after.TargetAt(from, t.DayType)
		if nt == nil {
			continue
		}
		if nt.Kcal < lim.KcalFloor {
			floorOK = false
			low = append(low, fmt.Sprintf("%s kcal", num(nt.Kcal)))
		}
		if t.DayType == DayLongRun {
			longRun = true
		}
	}
	if floorOK {
		checks = append(checks, Check{RuleKcalFloor, Pass, "kcal powyżej dolnego progu"})
	} else {
		checks = append(checks, Check{RuleKcalFloor, Fail, fmt.Sprintf("%s poniżej progu %s kcal", strings.Join(low, ", "), num(lim.KcalFloor))})
	}

	if longRun {
		nt := after.TargetAt(from, DayLongRun)
		switch {
		case lim.MaintenanceKcal == nil:
			checks = append(checks, Check{RuleLongRunDeficit, Warn, "brak szacunku zapotrzebowania, deficytu nie sprawdzono"})
		case nt.Kcal < *lim.MaintenanceKcal:
			checks = append(checks, Check{RuleLongRunDeficit, Fail,
				fmt.Sprintf("deficyt w dniu długiego: %s < %s kcal", num(nt.Kcal), num(*lim.MaintenanceKcal))})
		default:
			checks = append(checks, Check{RuleLongRunDeficit, Pass, "brak deficytu w dniu długiego"})
		}
	}

	if lim.WeightKg == nil || *lim.WeightKg <= 0 {
		checks = append(checks, Check{RuleProteinMin, Warn, "brak wagi, białka w g/kg nie sprawdzono"})
		return checks
	}
	minGKg := math.Inf(1)
	for _, t := range c.Targets {
		if nt := after.TargetAt(from, t.DayType); nt != nil {
			minGKg = math.Min(minGKg, nt.ProteinG / *lim.WeightKg)
		}
	}
	if minGKg < ProteinMinGKg {
		checks = append(checks, Check{RuleProteinMin, Warn, fmt.Sprintf("białko %s g/kg (minimum 1,6)", dec(minGKg, 2))})
	} else {
		checks = append(checks, Check{RuleProteinMin, Pass, "białko ≥ 1,6 g/kg"})
	}
	return checks
}

var hhmm = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

func validateReminder(r Reminder) Check {
	switch {
	case r.SupplementID == "":
		return Check{RuleReminderTime, Fail, "brak suplementu"}
	case !hhmm.MatchString(r.To) || !hhmm.MatchString(r.From):
		return Check{RuleReminderTime, Fail, "pora przypomnienia w formacie HH:MM"}
	case r.From == r.To:
		return Check{RuleReminderTime, Fail, "pora przypomnienia bez zmian"}
	}
	return Check{RuleReminderTime, Pass, fmt.Sprintf("przypomnienie %s → %s", r.From, r.To)}
}

var weekday = map[time.Weekday]string{
	time.Monday: "poniedziałek", time.Tuesday: "wtorek", time.Wednesday: "środa", time.Thursday: "czwartek",
	time.Friday: "piątek", time.Saturday: "sobota", time.Sunday: "niedziela",
}

func span(w time.Time) string {
	end := w.AddDate(0, 0, 6)
	return fmt.Sprintf("%d.%02d–%d.%02d", w.Day(), int(w.Month()), end.Day(), int(end.Month()))
}

// num formats a number the Polish way: no decimals when whole, else one, decimal comma.
func num(v float64) string {
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return dec(v, 1)
}

func dec(v float64, n int) string {
	return strings.Replace(strconv.FormatFloat(v, 'f', n, 64), ".", ",", 1)
}
