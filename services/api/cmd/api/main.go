// Command api serves the dashboard API and maintains projections built from raw_payload.
//
//	api serve    HTTP on :8080 (PORT), re-projects every PROJECT_INTERVAL (default 5m)
//	api project  rebuild all projections once and exit
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"forma/api/internal/dashboard"
	"forma/api/internal/httpapi"
	"forma/api/internal/plan"
	"forma/api/internal/projector"
	"forma/api/internal/syncreq"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("usage: api serve | api project")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("missing environment variable DATABASE_URL")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	switch os.Args[1] {
	case "project":
		stats, err := projector.Rebuild(ctx, pool)
		if err != nil {
			return err
		}
		if err := plan.NewStore(pool).RebuildProjections(ctx); err != nil {
			return err
		}
		slog.Info("projections rebuilt", "stats", stats)
		return nil
	case "serve":
		return serve(ctx, pool)
	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

func serve(ctx context.Context, pool *pgxpool.Pool) error {
	interval := 5 * time.Minute
	if v := os.Getenv("PROJECT_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("PROJECT_INTERVAL: %w", err)
		}
		interval = d
	}
	syncs := syncreq.NewStore(pool)
	go projectLoop(ctx, pool, syncs, interval)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.New(dashboard.NewService(dashboard.NewPgStore(pool), plan.NewStore(pool)), syncs),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	slog.Info("api listening", "addr", srv.Addr, "project_interval", interval)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// projectLoop rebuilds projections at startup, periodically and right after an on-demand sync
// (NOTIFY sync_done from the ingest scheduler), without coupling the Python workers to this service.
func projectLoop(ctx context.Context, pool *pgxpool.Pool, syncs *syncreq.Store, interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	synced := make(chan struct{}, 1)
	go syncreq.Listen(ctx, pool, func() {
		select {
		case synced <- struct{}{}:
		default: // a rebuild is already pending
		}
	}, func(err error) { slog.Error("listen sync_done failed", "err", err) })
	for {
		doneUpTo, err := syncs.DoneUpTo(ctx)
		if err != nil {
			slog.Error("sync_request read failed", "err", err)
		}
		if stats, err := projector.Rebuild(ctx, pool); err != nil {
			slog.Error("projection failed", "err", err)
		} else {
			syncs.MarkProjected(doneUpTo)
			slog.Info("projections rebuilt", "stats", stats)
		}
		// plan_session / plan_target are rewritten on every plan event; this only repairs them (e.g. after a seed).
		if err := plan.NewStore(pool).RebuildProjections(ctx); err != nil {
			slog.Error("plan projection failed", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		case <-synced:
			tick.Reset(interval)
		}
	}
}
