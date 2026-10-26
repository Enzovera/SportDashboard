// Package syncreq queues on-demand syncs from the dashboard ("Odśwież"). The API never talks to Garmin,
// Hevy or Fitatu: it inserts a sync_request row, the ingest scheduler claims it, runs the workers and
// marks it done with NOTIFY sync_done. Per-source results are the ingest_run rows started after the claim.
package syncreq

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Cooldown between the end of one request and the next: Garmin throttles frequent logins.
const Cooldown = 2 * time.Minute

// Channel the scheduler notifies when a request is done.
const Channel = "sync_done"

var ErrNotFound = errors.New("not found")

// TooSoonError means the previous request finished less than Cooldown ago.
type TooSoonError struct{ RetryAfter time.Duration }

func (e TooSoonError) Error() string {
	return fmt.Sprintf("sync finished recently, retry in %s", e.RetryAfter)
}

// SourceResult is the latest ingest run of one source within the request.
type SourceResult struct {
	Source string `json:"source"`
	Status string `json:"status"` // running | ok | partial | error
}

// Request is the body of POST /api/sync and GET /api/sync/{id}.
type Request struct {
	ID          int64          `json:"id"`
	Status      string         `json:"status"` // pending | running | done
	RequestedAt time.Time      `json:"requestedAt"`
	FinishedAt  *time.Time     `json:"finishedAt"`
	Sources     []SourceResult `json:"sources"`
}

// checkCooldown returns TooSoonError when lastFinished is within Cooldown of now.
func checkCooldown(lastFinished *time.Time, now time.Time) error {
	if lastFinished == nil {
		return nil
	}
	if left := lastFinished.Add(Cooldown).Sub(now); left > 0 {
		return TooSoonError{RetryAfter: left.Round(time.Second)}
	}
	return nil
}

// Store reports a request done only once projections were rebuilt after it (MarkProjected), so the
// dashboard reloads fresh data and not the state from before the sync.
type Store struct {
	pool      *pgxpool.Pool
	projected atomic.Int64 // highest done request covered by a finished projection rebuild
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// DoneUpTo is the highest finished request; read it before a rebuild and pass it to MarkProjected after.
func (s *Store) DoneUpTo(ctx context.Context) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `select coalesce(max(id), 0) from sync_request where status = 'done'`).Scan(&id)
	return id, err
}

func (s *Store) MarkProjected(upTo int64) {
	for {
		cur := s.projected.Load()
		if upTo <= cur || s.projected.CompareAndSwap(cur, upTo) {
			return
		}
	}
}

// Request returns the active request, or queues a new one. A second click joins the active request.
func (s *Store) Request(ctx context.Context) (Request, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `select id from sync_request where status in ('pending', 'running')`).Scan(&id)
	switch {
	case err == nil:
		return s.Get(ctx, id)
	case !errors.Is(err, pgx.ErrNoRows):
		return Request{}, fmt.Errorf("active sync_request: %w", err)
	}

	var last *time.Time
	if err := s.pool.QueryRow(ctx, `select max(finished_at) from sync_request`).Scan(&last); err != nil {
		return Request{}, fmt.Errorf("last sync_request: %w", err)
	}
	if err := checkCooldown(last, time.Now()); err != nil {
		return Request{}, err
	}

	// The partial unique index allows one active request; losing a race means joining the winner.
	err = s.pool.QueryRow(ctx, `
		insert into sync_request default values on conflict do nothing returning id`).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = s.pool.QueryRow(ctx, `select id from sync_request where status in ('pending', 'running')`).Scan(&id)
	}
	if err != nil {
		return Request{}, fmt.Errorf("insert sync_request: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *Store) Get(ctx context.Context, id int64) (Request, error) {
	var (
		r         = Request{ID: id, Sources: []SourceResult{}}
		startedAt *time.Time
	)
	err := s.pool.QueryRow(ctx, `
		select status, requested_at, started_at, finished_at from sync_request where id = $1`, id).
		Scan(&r.Status, &r.RequestedAt, &startedAt, &r.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("sync_request: %w", err)
	}
	if r.Status == "done" && id > s.projected.Load() {
		r.Status = "running" // ingest finished, projections not yet rebuilt
	}
	if startedAt == nil {
		return r, nil
	}
	rows, _ := s.pool.Query(ctx, `
		select distinct on (source) source, status
		from ingest_run
		where started_at >= $1 and ($2::timestamptz is null or started_at <= $2)
		order by source, started_at desc`, *startedAt, r.FinishedAt)
	if r.Sources, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (SourceResult, error) {
		var x SourceResult
		return x, row.Scan(&x.Source, &x.Status)
	}); err != nil {
		return Request{}, fmt.Errorf("ingest_run: %w", err)
	}
	return r, nil
}

// Listen calls onDone for every NOTIFY sync_done until ctx ends, reconnecting after errors.
func Listen(ctx context.Context, pool *pgxpool.Pool, onDone func(), onError func(error)) {
	for ctx.Err() == nil {
		if err := listenOnce(ctx, pool, onDone); err != nil && ctx.Err() == nil {
			onError(err)
			select {
			case <-ctx.Done():
			case <-time.After(10 * time.Second):
			}
		}
	}
}

func listenOnce(ctx context.Context, pool *pgxpool.Pool, onDone func()) error {
	pc, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	// LISTEN state stays on the connection, so it is taken out of the pool for good.
	conn := pc.Hijack()
	defer conn.Close(context.Background())
	if _, err := conn.Exec(ctx, "listen "+Channel); err != nil {
		return err
	}
	for {
		if _, err := conn.WaitForNotification(ctx); err != nil {
			return err
		}
		onDone()
	}
}
