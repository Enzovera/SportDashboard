// Package httpapi exposes the dashboard API. Routes:
//
//	GET /health
//	GET /api/today
//	GET /api/training?range=4w|12w|block
//	GET /api/body
//	GET /api/nutrition?range=14d
//
// /api/chat/* belongs to coach-chat (in dev Vite proxies it there directly; here it is not handled yet).
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"forma/api/internal/dashboard"
)

type Views interface {
	Today(ctx context.Context) (dashboard.TodayResponse, error)
	Training(ctx context.Context, rng string) (dashboard.TrainingResponse, error)
	Body(ctx context.Context) (dashboard.BodyResponse, error)
	Nutrition(ctx context.Context) (dashboard.NutritionResponse, error)
}

func New(v Views) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	mux.HandleFunc("GET /api/today", func(w http.ResponseWriter, r *http.Request) {
		resp, err := v.Today(r.Context())
		respond(w, r, resp, err)
	})
	mux.HandleFunc("GET /api/training", func(w http.ResponseWriter, r *http.Request) {
		rng := r.URL.Query().Get("range")
		switch rng {
		case "", "4w", "12w", "block":
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "range must be 4w, 12w or block"})
			return
		}
		resp, err := v.Training(r.Context(), rng)
		respond(w, r, resp, err)
	})
	mux.HandleFunc("GET /api/body", func(w http.ResponseWriter, r *http.Request) {
		resp, err := v.Body(r.Context())
		respond(w, r, resp, err)
	})
	mux.HandleFunc("GET /api/nutrition", func(w http.ResponseWriter, r *http.Request) {
		if rng := r.URL.Query().Get("range"); rng != "" && rng != "14d" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "range must be 14d"})
			return
		}
		resp, err := v.Nutrition(r.Context())
		respond(w, r, resp, err)
	})
	return mux
}

func respond(w http.ResponseWriter, r *http.Request, body any, err error) {
	if err != nil {
		slog.Error("request failed", "path", r.URL.Path, "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, body)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
