package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"forma/api/internal/dashboard"
)

// DayDetail is the full day_summary row; dashboard.DayRow carries only what the views need.
type DayDetail struct {
	Date        time.Time
	SleepS      *int
	SleepScore  *int
	HRVMs       *int
	HRVWeeklyMs *int
	HRVStatus   *string
	RestingHR   *int
	Steps       *int
	ActiveKcal  *int
	TotalKcal   *int
	WeightKg    *float64
}

type Note struct {
	ID        int64
	Date      time.Time
	Kind      string
	Content   string
	CreatedAt time.Time
}

// Store is read-only over projections except for coach notes; it deliberately does not embed dashboard.Store,
// which also writes (supplement log).
type Store interface {
	// Load returns projections from `since` onwards, plus sync status (dashboard.PgStore.Load).
	Load(ctx context.Context, since time.Time) (dashboard.Data, error)
	// Day returns the day_summary row for date, or nil when the day has no data.
	Day(ctx context.Context, date time.Time) (*DayDetail, error)
	// Notes returns coach notes dated `since` or later, newest first.
	Notes(ctx context.Context, since time.Time) ([]Note, error)
	SaveNote(ctx context.Context, date time.Time, kind, content string) (Note, error)
}

type PgStore struct {
	projections *dashboard.PgStore
	pool        *pgxpool.Pool
}

func NewPgStore(pool *pgxpool.Pool) *PgStore {
	return &PgStore{projections: dashboard.NewPgStore(pool), pool: pool}
}

func (s *PgStore) Load(ctx context.Context, since time.Time) (dashboard.Data, error) {
	return s.projections.Load(ctx, since)
}

func (s *PgStore) Day(ctx context.Context, date time.Time) (*DayDetail, error) {
	var x DayDetail
	err := s.pool.QueryRow(ctx, `
		select date, sleep_s, sleep_score, hrv_last_night_ms, hrv_weekly_avg_ms, hrv_status, resting_hr, steps,
		       active_kcal, total_kcal, weight_kg::float8
		from day_summary where date = $1`, date).
		Scan(&x.Date, &x.SleepS, &x.SleepScore, &x.HRVMs, &x.HRVWeeklyMs, &x.HRVStatus, &x.RestingHR, &x.Steps,
			&x.ActiveKcal, &x.TotalKcal, &x.WeightKg)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("day_summary: %w", err)
	}
	return &x, nil
}

func (s *PgStore) Notes(ctx context.Context, since time.Time) ([]Note, error) {
	rows, _ := s.pool.Query(ctx, `
		select id, date, kind, content, created_at from coach_note
		where date >= $1 order by date desc, id desc`, since)
	notes, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Note, error) {
		var n Note
		return n, r.Scan(&n.ID, &n.Date, &n.Kind, &n.Content, &n.CreatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("coach_note: %w", err)
	}
	return notes, nil
}

func (s *PgStore) SaveNote(ctx context.Context, date time.Time, kind, content string) (Note, error) {
	n := Note{Date: date, Kind: kind, Content: content}
	err := s.pool.QueryRow(ctx, `
		insert into coach_note (date, kind, content) values ($1, $2, $3) returning id, created_at`,
		date, kind, content).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return Note{}, fmt.Errorf("coach_note: %w", err)
	}
	return n, nil
}
