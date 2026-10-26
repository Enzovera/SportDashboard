package plan

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"forma/api/internal/projector"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid request")
	ErrNoPlan   = errors.New("no plan")
)

// MaxCommentLen is the longest decision comment accepted (HTTP 422 above it).
const MaxCommentLen = 500

// BlockedProposal is a plan_proposal_blocked row.
type BlockedProposal struct {
	ID         string
	Kind       string
	Origin     string
	Title      string
	Checks     []Check
	ProposedAt time.Time
}

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Current folds the current stream (the latest PlanCreated); nil without a plan.
func (s *Store) Current(ctx context.Context) (*State, error) { return current(ctx, s.pool) }

func current(ctx context.Context, q querier) (*State, error) {
	rows, _ := q.Query(ctx, `
		select stream_id::text, seq, type, payload, actor, created_at from plan_event
		where stream_id = (select stream_id from plan_event where type = 'PlanCreated' order by created_at desc, stream_id limit 1)
		order by seq`)
	events, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Event, error) {
		var e Event
		return e, r.Scan(&e.StreamID, &e.Seq, &e.Type, &e.Payload, &e.Actor, &e.CreatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("plan_event: %w", err)
	}
	return Fold(events)
}

// Blocked returns proposals a validator rejected since `since`, newest first.
func (s *Store) Blocked(ctx context.Context, since time.Time) ([]BlockedProposal, error) {
	rows, _ := s.pool.Query(ctx, `
		select id, kind, origin, title, checks, proposed_at from plan_proposal_blocked
		where proposed_at >= $1 order by proposed_at desc`, since)
	out, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (BlockedProposal, error) {
		var b BlockedProposal
		return b, r.Scan(&b.ID, &b.Kind, &b.Origin, &b.Title, &b.Checks, &b.ProposedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("plan_proposal_blocked: %w", err)
	}
	return out, nil
}

// Propose runs the validators on a coach proposal (MCP propose_plan_change). Passed: PlanChangeProposed is
// appended and the proposal waits for a human. Failed: it is stored as blocked only, never queued.
// The checks are returned either way so the coach gets the reason.
func (s *Store) Propose(ctx context.Context, p Proposed, lim Limits, now time.Time) ([]Check, bool, error) {
	if p.ProposalID == "" {
		p.ProposalID = newID()
	}
	p.V = PayloadVersion
	var checks []Check
	var blocked bool
	err := s.inTx(ctx, func(tx pgx.Tx, st *State) error {
		if st == nil {
			return ErrNoPlan
		}
		from := NewDate(now.In(projector.Warsaw)).Time
		if p.EffectiveFrom != nil {
			from = p.EffectiveFrom.Time
		}
		checks = Validate(st, p.Change, from, lim)
		p.Checks = checks
		payload, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if blocked = Blocked(checks); blocked {
			cj, _ := json.Marshal(checks)
			_, err := tx.Exec(ctx, `
				insert into plan_proposal_blocked (id, stream_id, kind, origin, title, checks, payload, proposed_at)
				values ($1, $2, $3, $4, $5, $6, $7, $8)`,
				p.ProposalID, st.StreamID, p.Kind, p.Origin, p.Title, cj, payload, now)
			return err
		}
		return appendEvent(ctx, tx, st, PlanChangeProposed, payload, "coach", now)
	})
	return checks, blocked, err
}

// Decide appends PlanChangeAccepted / PlanChangeRejected, applies an accepted reminder change and rewrites the
// projections, all in one transaction. ErrNotFound: unknown id; ErrConflict: already decided.
func (s *Store) Decide(ctx context.Context, id string, accept bool, comment *string, now time.Time) (Proposal, error) {
	if comment != nil && len([]rune(*comment)) > MaxCommentLen {
		return Proposal{}, fmt.Errorf("%w: comment longer than %d characters", ErrInvalid, MaxCommentLen)
	}
	var out Proposal
	err := s.inTx(ctx, func(tx pgx.Tx, st *State) error {
		if st == nil {
			return ErrNotFound
		}
		pr := st.Proposal(id)
		if pr == nil {
			return ErrNotFound
		}
		if pr.Decision != nil {
			return ErrConflict
		}
		typ := PlanChangeRejected
		if accept {
			typ = PlanChangeAccepted
		}
		payload, err := json.Marshal(Decided{V: PayloadVersion, ProposalID: id, Comment: comment})
		if err != nil {
			return err
		}
		if err := appendEvent(ctx, tx, st, typ, payload, "user", now); err != nil {
			return err
		}
		if r := pr.Change.Reminder; accept && r != nil {
			// The coach may move a reminder, nothing else: same supplement, same current time.
			if _, err := tx.Exec(ctx, `
				update supplement_rule set remind_at = $3::time
				where supplement_id = $1 and remind_at = $2::time`, r.SupplementID, r.From, r.To); err != nil {
				return fmt.Errorf("supplement_rule: %w", err)
			}
		}
		after, err := current(ctx, tx)
		if err != nil {
			return err
		}
		out = *after.Proposal(id)
		return writeProjections(ctx, tx, after)
	})
	return out, err
}

// RebuildProjections rewrites plan_session and plan_target from the events.
func (s *Store) RebuildProjections(ctx context.Context) error {
	return s.inTx(ctx, func(tx pgx.Tx, st *State) error { return writeProjections(ctx, tx, st) })
}

// inTx serializes plan writers (one stream, tiny traffic) and hands over the current state.
func (s *Store) inTx(ctx context.Context, fn func(pgx.Tx, *State) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit
	if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtext('plan_event'))`); err != nil {
		return err
	}
	st, err := current(ctx, tx)
	if err != nil {
		return err
	}
	if err := fn(tx, st); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func appendEvent(ctx context.Context, tx pgx.Tx, st *State, typ string, payload []byte, actor string, now time.Time) error {
	_, err := tx.Exec(ctx, `
		insert into plan_event (stream_id, seq, type, payload, actor, created_at) values ($1, $2, $3, $4, $5, $6)`,
		st.StreamID, st.LastSeq+1, typ, payload, actor, now)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation: concurrent append or second decision
		return ErrConflict
	}
	return err
}

func writeProjections(ctx context.Context, tx pgx.Tx, st *State) error {
	if _, err := tx.Exec(ctx, `delete from plan_session; delete from plan_target`); err != nil {
		return fmt.Errorf("clear plan projections: %w", err)
	}
	if st == nil {
		return nil
	}
	var sessions, targets [][]any
	for _, d := range st.Days {
		for i, x := range d.Sessions {
			sessions = append(sessions, []any{d.Date.Time, i + 1, x.Kind, x.Title, x.Detail, x.DistanceKm, d.DayType})
		}
	}
	for _, t := range st.Targets {
		targets = append(targets, []any{t.From, t.DayType, int(t.Kcal), int(t.ProteinG), int(t.CarbsG), int(t.FatG)})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"plan_session"},
		[]string{"date", "slot", "kind", "title", "detail", "distance_km", "day_type"}, pgx.CopyFromRows(sessions)); err != nil {
		return fmt.Errorf("copy plan_session: %w", err)
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"plan_target"},
		[]string{"from_date", "day_type", "kcal", "protein_g", "carbs_g", "fat_g"}, pgx.CopyFromRows(targets)); err != nil {
		return fmt.Errorf("copy plan_target: %w", err)
	}
	return nil
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "p-" + hex.EncodeToString(b)
}
