package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"forma/api/internal/syncreq"
)

type fakeSyncer struct {
	req    syncreq.Request
	reqErr error
	getErr error
	gotID  int64
}

func (f *fakeSyncer) Request(context.Context) (syncreq.Request, error) { return f.req, f.reqErr }

func (f *fakeSyncer) Get(_ context.Context, id int64) (syncreq.Request, error) {
	f.gotID = id
	return f.req, f.getErr
}

func TestSyncRoutes(t *testing.T) {
	done := syncreq.Request{ID: 7, Status: "done", Sources: []syncreq.SourceResult{{Source: "hevy", Status: "ok"}}}
	tests := []struct {
		name       string
		method     string
		path       string
		syncer     fakeSyncer
		wantStatus int
		wantRetry  string
		wantID     int64 // id passed to Get
	}{
		{"queue", "POST", "/api/sync", fakeSyncer{req: syncreq.Request{ID: 7, Status: "pending"}}, http.StatusAccepted, "", 0},
		{"cooldown", "POST", "/api/sync", fakeSyncer{reqErr: syncreq.TooSoonError{RetryAfter: 90 * time.Second}}, http.StatusTooManyRequests, "90", 0},
		{"queue fails", "POST", "/api/sync", fakeSyncer{reqErr: errors.New("db down")}, http.StatusInternalServerError, "", 0},
		{"status", "GET", "/api/sync/7", fakeSyncer{req: done}, http.StatusOK, "", 7},
		{"unknown id", "GET", "/api/sync/8", fakeSyncer{getErr: syncreq.ErrNotFound}, http.StatusNotFound, "", 8},
		{"bad id", "GET", "/api/sync/abc", fakeSyncer{}, http.StatusNotFound, "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.syncer
			rec := httptest.NewRecorder()
			New(nil, &s).ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if got := rec.Header().Get("Retry-After"); got != tt.wantRetry {
				t.Errorf("Retry-After = %q, want %q", got, tt.wantRetry)
			}
			if s.gotID != tt.wantID {
				t.Errorf("Get id = %d, want %d", s.gotID, tt.wantID)
			}
			if rec.Code < 300 {
				var body syncreq.Request
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.ID != 7 {
					t.Errorf("body = %s (%v)", rec.Body, err)
				}
			}
		})
	}
}
