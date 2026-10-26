package dashboard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"forma/api/internal/plan"
)

// Domain errors, shared with the plan store so the HTTP layer maps them once:
// ErrNotFound (404), ErrConflict (409), ErrInvalid (422).
var (
	ErrNotFound = plan.ErrNotFound
	ErrConflict = plan.ErrConflict
	ErrInvalid  = plan.ErrInvalid
)

type SupplementRuleRow struct {
	When     string
	Slot     string
	RemindAt *string // "HH:MM"
	Note     *string
}

type SupplementRow struct {
	ID              string
	Name            string
	DoseAmount      float64
	DoseUnit        string
	PackUnit        string
	PackPerDose     float64
	CountedInFitatu bool
	ActiveFrom      time.Time
	ActiveTo        *time.Time
	Rules           []SupplementRuleRow // ordered by id
}

type SupplementLogRow struct {
	SupplementID string
	Date         time.Time
	Type         string // SupplementTaken | SupplementSkipped
	CreatedAt    time.Time
}

type StockRow struct {
	SupplementID string
	Amount       float64
	CreatedAt    time.Time
}

type SupplementData struct {
	Supplements []SupplementRow // in display order
	Logs        []SupplementLogRow
	Stock       []StockRow
}

func (s *PgStore) LoadSupplements(ctx context.Context, since time.Time) (SupplementData, error) {
	var d SupplementData
	rows, err := s.pool.Query(ctx, `
		select s.id, s.name, s.dose_amount::float8, s.dose_unit, s.pack_unit, s.pack_per_dose::float8,
		       s.counted_in_fitatu, s.active_from, s.active_to,
		       r."when", r.slot, to_char(r.remind_at, 'HH24:MI'), r.note
		from supplement s left join supplement_rule r on r.supplement_id = s.id
		order by s.sort, s.id, r.id`)
	if err != nil {
		return d, fmt.Errorf("supplement: %w", err)
	}
	for rows.Next() {
		var (
			x          SupplementRow
			when, slot *string
			rule       SupplementRuleRow
		)
		if err := rows.Scan(&x.ID, &x.Name, &x.DoseAmount, &x.DoseUnit, &x.PackUnit, &x.PackPerDose, &x.CountedInFitatu,
			&x.ActiveFrom, &x.ActiveTo, &when, &slot, &rule.RemindAt, &rule.Note); err != nil {
			return d, fmt.Errorf("supplement: %w", err)
		}
		if n := len(d.Supplements); n == 0 || d.Supplements[n-1].ID != x.ID {
			d.Supplements = append(d.Supplements, x)
		}
		if when != nil {
			rule.When, rule.Slot = *when, *slot
			last := &d.Supplements[len(d.Supplements)-1]
			last.Rules = append(last.Rules, rule)
		}
	}
	if err := rows.Err(); err != nil {
		return d, fmt.Errorf("supplement: %w", err)
	}

	// Stock consumption needs the log since the latest stock entry, which can be older than `since`.
	rows, _ = s.pool.Query(ctx, `
		select supplement_id, date, type, created_at from supplement_log
		where date >= least($1, coalesce((select min(created_at)::date from supplement_stock), $1))
		order by id`, since)
	if d.Logs, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (SupplementLogRow, error) {
		var x SupplementLogRow
		return x, r.Scan(&x.SupplementID, &x.Date, &x.Type, &x.CreatedAt)
	}); err != nil {
		return d, fmt.Errorf("supplement_log: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `
		select distinct on (supplement_id) supplement_id, amount::float8, created_at
		from supplement_stock order by supplement_id, created_at desc, id desc`)
	if d.Stock, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (StockRow, error) {
		var x StockRow
		return x, r.Scan(&x.SupplementID, &x.Amount, &x.CreatedAt)
	}); err != nil {
		return d, fmt.Errorf("supplement_stock: %w", err)
	}
	return d, nil
}

// AppendSupplementLog appends SupplementTaken / SupplementSkipped; ErrNotFound for an unknown supplement.
func (s *PgStore) AppendSupplementLog(ctx context.Context, supplementID string, date time.Time, typ string) error {
	_, err := s.pool.Exec(ctx, `insert into supplement_log (supplement_id, date, type) values ($1, $2, $3)`,
		supplementID, date, typ)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
		return ErrNotFound
	}
	return err
}
