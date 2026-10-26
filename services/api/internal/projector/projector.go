// Package projector rebuilds projections (day_summary, activity, workout, strength_set, body_composition,
// nutrition_day, nutrition_meal) from raw_payload.
//
// Every run is a full rebuild inside one transaction: projections are disposable by design
// (docs/design.md, "Model danych"), the data is small, and a full rebuild is trivially correct after a
// parser change. Only the newest version of each record (highest raw_payload.id) is projected.
package projector

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Stats struct {
	Days       int `json:"days"`
	Activities int `json:"activities"`
	Workouts   int `json:"workouts"`
	Sets       int `json:"sets"`
	WeighIns   int `json:"weighIns"`
	Nutrition  int `json:"nutritionDays"`
	Skipped    int `json:"skipped"`
}

type payload struct {
	id         int64
	source     string
	endpoint   string
	externalID string
	forDate    *time.Time
	body       []byte
}

// Projection is everything one rebuild writes.
type Projection struct {
	Days       []Day
	Activities []Activity
	Workouts   []Workout
	WeighIns   []WeighIn
	Nutrition  []NutritionDay
	Skipped    int
}

// Build turns the newest payloads into projection rows. Pure: no I/O, so it is unit-tested directly.
// A payload that fails to parse is skipped and counted, never fatal: one odd record must not block the dashboard.
func Build(latest []payload) (Projection, []error) {
	var (
		p       Projection
		errs    []error
		days    = map[time.Time]*Day{}
		zones   = map[string][]float64{}
		muscles = map[string]*string{}
		deleted = map[string]int64{} // workout id -> payload id of the delete event
		weighIn = map[time.Time]WeighIn{}
		diet    []payload
		targets = map[time.Time]*DietTarget{}
		acts    []payload
		works   []payload
	)
	day := func(t time.Time) *Day {
		if d, ok := days[t]; ok {
			return d
		}
		d := &Day{Date: t}
		days[t] = d
		return d
	}
	skip := func(err error) {
		p.Skipped++
		errs = append(errs, err)
	}

	for _, r := range latest {
		key := r.source + "/" + r.endpoint
		switch key {
		case "garmin/sleep", "garmin/hrv", "garmin/user_summary", "garmin/body_composition":
			if r.forDate == nil {
				skip(fmt.Errorf("%s %s: no for_date", key, r.externalID))
				continue
			}
			apply := map[string]func(*Day, []byte) error{
				"garmin/sleep": ApplySleep, "garmin/hrv": ApplyHRV,
				"garmin/user_summary": ApplyUserSummary, "garmin/body_composition": ApplyBodyComposition,
			}[key]
			if err := apply(day(*r.forDate), r.body); err != nil {
				skip(err)
				continue
			}
			if key == "garmin/body_composition" {
				ws, err := ParseWeighIns(r.body)
				if err != nil {
					skip(err)
					continue
				}
				for _, w := range ws {
					weighIn[w.MeasuredAt] = w // the same weigh-in can come in the payloads of two dates
				}
			}
		case "garmin/activity_hr_zones":
			z, err := ParseZones(r.body)
			if err != nil {
				skip(err)
				continue
			}
			zones[r.externalID] = z
		case "garmin/activities":
			acts = append(acts, r)
		case "hevy/exercise_templates":
			id, muscle, err := ParseTemplate(r.body)
			if err != nil {
				skip(err)
				continue
			}
			muscles[id] = muscle
		case "hevy/workouts":
			works = append(works, r)
		case "hevy/workout_deleted":
			deleted[r.externalID] = r.id
		case "fitatu/diet_day", "fitatu/settings":
			if r.forDate == nil {
				skip(fmt.Errorf("%s %s: no for_date", key, r.externalID))
				continue
			}
			if key == "fitatu/diet_day" {
				diet = append(diet, r)
				continue
			}
			t, err := ParseSettings(r.body)
			if err != nil {
				skip(err)
				continue
			}
			if t != nil {
				targets[*r.forDate] = t
			}
		}
	}

	for _, r := range acts {
		a, err := ParseActivity(r.body)
		if err != nil {
			skip(err)
			continue
		}
		a.HRZonesS = zones[a.ExternalID]
		p.Activities = append(p.Activities, a)
	}
	for _, r := range works {
		// Deleted after its last update: gone from Hevy, gone from the dashboard.
		if delID, ok := deleted[r.externalID]; ok && delID > r.id {
			continue
		}
		w, err := ParseWorkout(r.body, muscles)
		if err != nil {
			skip(err)
			continue
		}
		p.Workouts = append(p.Workouts, w)
	}
	for _, d := range days {
		p.Days = append(p.Days, *d)
	}
	for _, r := range diet {
		n, ok, err := ParseDietDay(*r.forDate, r.body)
		if err != nil {
			skip(err)
			continue
		}
		if ok {
			n.Target = targetFor(targets, n.Date)
			p.Nutrition = append(p.Nutrition, n)
		}
	}
	sort.Slice(p.Nutrition, func(i, j int) bool { return p.Nutrition[i].Date.Before(p.Nutrition[j].Date) })
	for _, w := range weighIn {
		p.WeighIns = append(p.WeighIns, w)
	}
	sort.Slice(p.WeighIns, func(i, j int) bool { return p.WeighIns[i].MeasuredAt.Before(p.WeighIns[j].MeasuredAt) })
	sort.Slice(p.Days, func(i, j int) bool { return p.Days[i].Date.Before(p.Days[j].Date) })
	sort.Slice(p.Activities, func(i, j int) bool { return p.Activities[i].StartedAt.Before(p.Activities[j].StartedAt) })
	sort.Slice(p.Workouts, func(i, j int) bool { return p.Workouts[i].StartedAt.Before(p.Workouts[j].StartedAt) })
	return p, errs
}

const latestSQL = `
select distinct on (source, endpoint, coalesce(external_id, ''))
       id, source, endpoint, coalesce(external_id, ''), for_date, body
from raw_payload
where source in ('garmin', 'hevy', 'fitatu')
order by source, endpoint, coalesce(external_id, ''), id desc`

// Rebuild replaces all projections with a fresh build from raw_payload.
func Rebuild(ctx context.Context, pool *pgxpool.Pool) (Stats, error) {
	rows, err := pool.Query(ctx, latestSQL)
	if err != nil {
		return Stats{}, fmt.Errorf("load raw_payload: %w", err)
	}
	latest, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (payload, error) {
		var p payload
		err := row.Scan(&p.id, &p.source, &p.endpoint, &p.externalID, &p.forDate, &p.body)
		return p, err
	})
	if err != nil {
		return Stats{}, fmt.Errorf("scan raw_payload: %w", err)
	}

	proj, errs := Build(latest)
	for i, e := range errs {
		if i == 5 {
			slog.Warn("more payloads skipped", "count", len(errs)-i)
			break
		}
		slog.Warn("payload skipped", "err", e)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return Stats{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit
	if err := write(ctx, tx, proj); err != nil {
		return Stats{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Stats{}, err
	}

	st := Stats{Days: len(proj.Days), Activities: len(proj.Activities), Workouts: len(proj.Workouts),
		WeighIns: len(proj.WeighIns), Nutrition: len(proj.Nutrition), Skipped: proj.Skipped}
	for _, w := range proj.Workouts {
		st.Sets += len(w.Sets)
	}
	return st, nil
}

func write(ctx context.Context, tx pgx.Tx, p Projection) error {
	if _, err := tx.Exec(ctx, `truncate day_summary, activity, strength_set, workout, body_composition, nutrition_day, nutrition_meal`); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}

	dayRows := make([][]any, 0, len(p.Days))
	for _, d := range p.Days {
		dayRows = append(dayRows, []any{d.Date, d.SleepS, d.SleepScore, d.HRVLastNightMs, d.HRVWeeklyAvgMs, d.HRVStatus,
			d.RestingHR, d.Steps, d.ActiveKcal, d.TotalKcal, d.WeightKg})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"day_summary"},
		[]string{"date", "sleep_s", "sleep_score", "hrv_last_night_ms", "hrv_weekly_avg_ms", "hrv_status",
			"resting_hr", "steps", "active_kcal", "total_kcal", "weight_kg"},
		pgx.CopyFromRows(dayRows)); err != nil {
		return fmt.Errorf("copy day_summary: %w", err)
	}

	actRows := make([][]any, 0, len(p.Activities))
	for _, a := range p.Activities {
		var zones any
		if a.HRZonesS != nil {
			zones = a.HRZonesS
		}
		actRows = append(actRows, []any{"garmin", a.ExternalID, a.StartedAt, a.LocalDate, a.Sport, a.Name, a.DistanceM,
			a.DurationS, a.MovingS, a.AvgHR, a.MaxHR, a.ElevationGainM, a.TrainingLoad, zones})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"activity"},
		[]string{"source", "external_id", "started_at", "local_date", "sport", "name", "distance_m", "duration_s",
			"moving_s", "avg_hr", "max_hr", "elevation_gain_m", "training_load", "hr_zones_s"},
		pgx.CopyFromRows(actRows)); err != nil {
		return fmt.Errorf("copy activity: %w", err)
	}

	var workRows, setRows [][]any
	for _, w := range p.Workouts {
		workRows = append(workRows, []any{w.ID, w.Title, w.StartedAt, w.EndedAt, w.LocalDate, w.VolumeKg})
		for _, s := range w.Sets {
			setRows = append(setRows, []any{w.ID, s.ExerciseIndex, s.SetNo, s.ExerciseTemplateID, s.Exercise, s.MuscleGroup,
				s.SetType, s.WeightKg, s.Reps, s.RPE, s.DistanceM, s.DurationS})
		}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"workout"},
		[]string{"id", "title", "started_at", "ended_at", "local_date", "volume_kg"}, pgx.CopyFromRows(workRows)); err != nil {
		return fmt.Errorf("copy workout: %w", err)
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"strength_set"},
		[]string{"workout_id", "exercise_index", "set_no", "exercise_template_id", "exercise", "muscle_group",
			"set_type", "weight_kg", "reps", "rpe", "distance_m", "duration_s"}, pgx.CopyFromRows(setRows)); err != nil {
		return fmt.Errorf("copy strength_set: %w", err)
	}

	weighRows := make([][]any, 0, len(p.WeighIns))
	for _, w := range p.WeighIns {
		weighRows = append(weighRows, []any{w.MeasuredAt, w.LocalDate, w.WeightKg, w.FatPct, w.MuscleKg, w.WaterPct, w.SourceType})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"body_composition"},
		[]string{"measured_at", "local_date", "weight_kg", "fat_pct", "muscle_kg", "water_pct", "source_type"},
		pgx.CopyFromRows(weighRows)); err != nil {
		return fmt.Errorf("copy body_composition: %w", err)
	}

	var nutRows, mealRows [][]any
	for _, n := range p.Nutrition {
		row := []any{n.Date, n.Kcal, n.ProteinG, n.CarbsG, n.FatG, nil, nil, nil, nil}
		if t := n.Target; t != nil {
			row[5], row[6], row[7], row[8] = t.Kcal, t.ProteinG, t.CarbsG, t.FatG
		}
		nutRows = append(nutRows, row)
		for _, m := range n.Meals {
			mealRows = append(mealRows, []any{n.Date, m.Slot, m.SlotNo, m.At, m.Title, m.Kcal, m.ProteinG, m.CarbsG, m.FatG})
		}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"nutrition_day"},
		[]string{"date", "kcal", "protein_g", "carbs_g", "fat_g", "kcal_target", "protein_target_g", "carbs_target_g", "fat_target_g"},
		pgx.CopyFromRows(nutRows)); err != nil {
		return fmt.Errorf("copy nutrition_day: %w", err)
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"nutrition_meal"},
		[]string{"date", "slot", "slot_no", "at", "title", "kcal", "protein_g", "carbs_g", "fat_g"},
		pgx.CopyFromRows(mealRows)); err != nil {
		return fmt.Errorf("copy nutrition_meal: %w", err)
	}
	return nil
}
