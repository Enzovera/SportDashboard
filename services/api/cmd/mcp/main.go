// Command mcp serves the forma MCP server over streamable HTTP on :8091 (PORT) at /mcp.
// Needs DATABASE_URL and FORMA_MCP_TOKEN (clients send `Authorization: Bearer <token>`).
// It reads projections maintained by `api serve`; it never rebuilds them itself.
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

	"forma/api/internal/mcpserver"
)

func main() {
	if err := run(); err != nil {
		slog.Error("mcp failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn, token := os.Getenv("DATABASE_URL"), os.Getenv("FORMA_MCP_TOKEN")
	if dsn == "" || token == "" {
		return errors.New("missing environment variable DATABASE_URL or FORMA_MCP_TOKEN")
	}
	if len(token) < 32 {
		return errors.New("FORMA_MCP_TOKEN must have at least 32 characters")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.Handle("/mcp", mcpserver.Handler(mcpserver.New(mcpserver.NewPgStore(pool), time.Now), token))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}
	srv := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	slog.Info("mcp listening", "addr", srv.Addr, "version", mcpserver.Version)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
