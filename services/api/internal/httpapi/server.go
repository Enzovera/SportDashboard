// Package httpapi exposes the dashboard API. Routes:
//
//	GET /health
//	GET /api/today
//	GET /api/training?range=4w|12w|block
//	GET /api/body
//	GET /api/nutrition?range=14d
//	GET /api/supplements
//	POST /api/supplements/{id}/log   body {status: taken|skipped, date}
//	GET /api/plan
//	GET /api/proposals
//	POST /api/proposals/{id}/accept|reject   body {comment}   404 unknown, 409 decided, 422 comment > 500
//
// /api/chat/* belongs to coach-chat (in dev Vite proxies it there directly; here it is not handled yet).
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"forma/api/internal/dashboard"
)

type Views interface {
	Today(ctx context.Context) (dashboard.TodayResponse, error)
	Training(ctx context.Context, rng string) (dashboard.TrainingResponse, error)
	Body(ctx context.Context) (dashboard.BodyResponse, error)
	Nutrition(ctx context.Context) (dashboard.NutritionResponse, error)
	Supplements(ctx context.Context) (dashboard.SupplementsResponse, error)
	LogSupplement(ctx context.Context, id, status, date string) (dashboard.SupplementItem, error)
	Plan(ctx context.Context) (dashboard.PlanResponse, error)
	Proposals(ctx context.Context) (dashboard.ProposalsResponse, error)
	Decide(ctx context.Context, id string, accept bool, comment *string) (dashboard.ProposalDecision, error)
}

// maxBody caps request bodies; every POST body here is a few short fields.
const maxBody = 4 << 10

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
	mux.HandleFunc("GET /api/supplements", func(w http.ResponseWriter, r *http.Request) {
		resp, err := v.Supplements(r.Context())
		respond(w, r, resp, err)
	})
	mux.HandleFunc("POST /api/supplements/{id}/log", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Status string `json:"status"`
			Date   string `json:"date"`
		}
		if !decode(w, r, &body) {
			return
		}
		resp, err := v.LogSupplement(r.Context(), r.PathValue("id"), body.Status, body.Date)
		respond(w, r, resp, err)
	})
	mux.HandleFunc("GET /api/plan", func(w http.ResponseWriter, r *http.Request) {
		resp, err := v.Plan(r.Context())
		respond(w, r, resp, err)
	})
	mux.HandleFunc("GET /api/proposals", func(w http.ResponseWriter, r *http.Request) {
		resp, err := v.Proposals(r.Context())
		respond(w, r, resp, err)
	})
	decide := func(accept bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Comment *string `json:"comment"`
			}
			if !decode(w, r, &body) {
				return
			}
			resp, err := v.Decide(r.Context(), r.PathValue("id"), accept, body.Comment)
			respond(w, r, resp, err)
		}
	}
	mux.HandleFunc("POST /api/proposals/{id}/accept", decide(true))
	mux.HandleFunc("POST /api/proposals/{id}/reject", decide(false))
	return mux
}

// decode reads a JSON body into dst; on failure it writes 400 and returns false.
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return false
	}
	return true
}

// respond maps domain errors to HTTP statuses; anything unexpected is logged and hidden behind a 500.
func respond(w http.ResponseWriter, r *http.Request, body any, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, body)
	case errors.Is(err, dashboard.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, dashboard.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already decided"})
	case errors.Is(err, dashboard.ErrInvalid):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	default:
		slog.Error("request failed", "path", r.URL.Path, "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
