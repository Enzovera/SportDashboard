package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Rows read from projections. Everything the views need comes from here; no raw_payload access.

type DayRow struct {
	Date      time.Time
	SleepS    *int
	HRVMs     *float64 // last night average
	RestingHR *int
	WeightKg  *float64
}

type ActivityRow struct {
	ID        string
	StartedAt time.Time
	LocalDate time.Time
	Sport     string
	Name      *string
	DistanceM *float64
	DurationS *float64
	AvgHR     *int
	Load      *float64
}

type WorkoutRow struct {
	ID          string
	Title       *string
	LocalDate   time.Time
	StartedAt   time.Time
	EndedAt     *time.Time
	WorkingSets int
	TonnageKg   float64
	Muscles     map[string]int // primary muscle group -> working sets
}

type SyncRow struct {
	Source        string
	LastStatus    string
	LastSuccessAt *time.Time
}

type WeighInRow struct {
	MeasuredAt time.Time
	LocalDate  time.Time
	WeightKg   float64
	FatPct     *float64
	MuscleKg   *float64
	WaterPct   *float64
}

// DietTarget is a daily nutrition goal (Fitatu settings now, plan per day type in stage 4).
type DietTarget struct {
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
}

type NutritionDayRow struct {
	Date     time.Time
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
	Target   *DietTarget // Fitatu goal in force that day
}

type MealRow struct {
	Date     time.Time
	Slot     string
	At       *time.Time
	Title    string
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
}

type Data struct {
	Days       []DayRow
	Activities []ActivityRow
	Workouts   []WorkoutRow
	WeighIns   []WeighInRow      // oldest first
	Nutrition  []NutritionDayRow // days with a logged meal, oldest first
	Meals      []MealRow         // in Fitatu slot order
	Sync       []SyncRow
}

type Store interface {
	// Load returns projections from `since` (local calendar date) onwards, plus sync status.
	Load(ctx context.Context, since time.Time) (Data, error)
	// LoadSupplements returns the supplement set, rules, log from `since` and the latest stock counts.
	LoadSupplements(ctx context.Context, since time.Time) (SupplementData, error)
	AppendSupplementLog(ctx context.Context, supplementID string, date time.Time, typ string) error
}

type PgStore struct{ pool *pgxpool.Pool }

func NewPgStore(pool *pgxpool.Pool) *PgStore { return &PgStore{pool: pool} }

func (s *PgStore) Load(ctx context.Context, since time.Time) (Data, error) {
	var d Data
	var err error

	rows, _ := s.pool.Query(ctx, `
		select date, sleep_s, hrv_last_night_ms::float8, resting_hr, weight_kg::float8
		from day_summary where date >= $1 order by date`, since)
	if d.Days, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (DayRow, error) {
		var x DayRow
		return x, r.Scan(&x.Date, &x.SleepS, &x.HRVMs, &x.RestingHR, &x.WeightKg)
	}); err != nil {
		return d, fmt.Errorf("day_summary: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `
		select external_id, started_at, local_date, sport, name, distance_m::float8, duration_s::float8, avg_hr,
		       training_load::float8
		from activity where local_date >= $1 order by started_at`, since)
	if d.Activities, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (ActivityRow, error) {
		var x ActivityRow
		return x, r.Scan(&x.ID, &x.StartedAt, &x.LocalDate, &x.Sport, &x.Name, &x.DistanceM, &x.DurationS, &x.AvgHR, &x.Load)
	}); err != nil {
		return d, fmt.Errorf("activity: %w", err)
	}

	// One row per (workout, muscle group): working sets (no warm-ups, reps done) and tonnage.
	byID := map[string]*WorkoutRow{}
	var order []string
	rows, err = s.pool.Query(ctx, `
		select w.id, w.title, w.local_date, w.started_at, w.ended_at, s.muscle_group,
		       count(s.*) filter (where s.set_type <> 'warmup' and s.reps > 0)::int,
		       coalesce(sum(s.weight_kg * s.reps) filter (where s.set_type <> 'warmup' and s.reps > 0), 0)::float8
		from workout w left join strength_set s on s.workout_id = w.id
		where w.local_date >= $1
		group by w.id, s.muscle_group
		order by w.started_at`, since)
	if err != nil {
		return d, fmt.Errorf("workout: %w", err)
	}
	for rows.Next() {
		var (
			x       WorkoutRow
			muscle  *string
			sets    int
			tonnage float64
		)
		if err := rows.Scan(&x.ID, &x.Title, &x.LocalDate, &x.StartedAt, &x.EndedAt, &muscle, &sets, &tonnage); err != nil {
			return d, fmt.Errorf("workout: %w", err)
		}
		w, ok := byID[x.ID]
		if !ok {
			x.Muscles = map[string]int{}
			w = &x
			byID[x.ID] = w
			order = append(order, x.ID)
		}
		w.WorkingSets += sets
		w.TonnageKg += tonnage
		if muscle != nil && sets > 0 {
			w.Muscles[*muscle] += sets
		}
	}
	if err := rows.Err(); err != nil {
		return d, fmt.Errorf("workout: %w", err)
	}
	for _, id := range order {
		d.Workouts = append(d.Workouts, *byID[id])
	}

	rows, _ = s.pool.Query(ctx, `
		select measured_at, local_date, weight_kg::float8, fat_pct::float8, muscle_kg::float8, water_pct::float8
		from body_composition where local_date >= $1 order by measured_at`, since)
	if d.WeighIns, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (WeighInRow, error) {
		var x WeighInRow
		return x, r.Scan(&x.MeasuredAt, &x.LocalDate, &x.WeightKg, &x.FatPct, &x.MuscleKg, &x.WaterPct)
	}); err != nil {
		return d, fmt.Errorf("body_composition: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `
		select date, kcal::float8, protein_g::float8, carbs_g::float8, fat_g::float8,
		       kcal_target::float8, protein_target_g::float8, carbs_target_g::float8, fat_target_g::float8
		from nutrition_day where date >= $1 order by date`, since)
	if d.Nutrition, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (NutritionDayRow, error) {
		var (
			x       NutritionDayRow
			kcal    *float64
			p, c, f *float64
		)
		if err := r.Scan(&x.Date, &x.Kcal, &x.ProteinG, &x.CarbsG, &x.FatG, &kcal, &p, &c, &f); err != nil {
			return x, err
		}
		if kcal != nil {
			x.Target = &DietTarget{Kcal: *kcal, ProteinG: deref(p), CarbsG: deref(c), FatG: deref(f)}
		}
		return x, nil
	}); err != nil {
		return d, fmt.Errorf("nutrition_day: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `
		select date, slot, at, title, kcal::float8, protein_g::float8, carbs_g::float8, fat_g::float8
		from nutrition_meal where date >= $1 order by date, slot_no`, since)
	if d.Meals, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (MealRow, error) {
		var x MealRow
		return x, r.Scan(&x.Date, &x.Slot, &x.At, &x.Title, &x.Kcal, &x.ProteinG, &x.CarbsG, &x.FatG)
	}); err != nil {
		return d, fmt.Errorf("nutrition_meal: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `
		select source,
		       (array_agg(status order by started_at desc))[1],
		       max(finished_at) filter (where status in ('ok', 'partial'))
		from ingest_run group by source order by source`)
	if d.Sync, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (SyncRow, error) {
		var x SyncRow
		return x, r.Scan(&x.Source, &x.LastStatus, &x.LastSuccessAt)
	}); err != nil {
		return d, fmt.Errorf("ingest_run: %w", err)
	}
	return d, nil
}

func deref(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
