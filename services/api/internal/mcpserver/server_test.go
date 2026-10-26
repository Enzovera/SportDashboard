package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"forma/api/internal/dashboard"
)

type fakeStore struct {
	data  dashboard.Data
	notes []Note
}

func (f *fakeStore) Load(context.Context, time.Time) (dashboard.Data, error) { return f.data, nil }

func (f *fakeStore) Day(context.Context, time.Time) (*DayDetail, error) { return nil, nil }

func (f *fakeStore) Notes(_ context.Context, since time.Time) ([]Note, error) {
	var out []Note
	for i := len(f.notes) - 1; i >= 0; i-- {
		if !f.notes[i].Date.Before(since) {
			out = append(out, f.notes[i])
		}
	}
	return out, nil
}

func (f *fakeStore) SaveNote(_ context.Context, date time.Time, kind, content string) (Note, error) {
	n := Note{ID: int64(len(f.notes) + 1), Date: date, Kind: kind, Content: content, CreatedAt: now}
	f.notes = append(f.notes, n)
	return n, nil
}

func connect(t *testing.T, store Store) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	if _, err := New(store, func() time.Time { return now }).Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any, out any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if out != nil && !res.IsError {
		b, _ := json.Marshal(res.StructuredContent)
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatal(err)
		}
	}
	return res
}

func TestToolList(t *testing.T) {
	cs := connect(t, &fakeStore{})
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
		if tool.Name == "get_trends" {
			b, _ := json.Marshal(tool.InputSchema)
			if !strings.Contains(string(b), `"run_km_week"`) || !strings.Contains(string(b), `"required":["metric"]`) {
				t.Errorf("get_trends schema: %s", b)
			}
		}
	}
	sort.Strings(names)
	want := "get_activities,get_coach_log,get_day_summary,get_strength_volume,get_trends,save_coach_note"
	if strings.Join(names, ",") != want {
		t.Errorf("tools = %v", names)
	}
	for _, n := range names {
		if strings.Contains(n, "accept") || strings.Contains(n, "delete") {
			t.Errorf("tool %s must not exist", n)
		}
	}
}

func TestTrendsTool(t *testing.T) {
	cs := connect(t, &fakeStore{data: sample()})
	var got Trend
	call(t, cs, "get_trends", map[string]any{"metric": "run_km_week", "from": "2026-10-19"}, &got)
	if len(got.Points) != 1 || *got.Points[0].Value != 25 || got.Unit != "km" {
		t.Errorf("trend: %+v", got)
	}
	if res := call(t, cs, "get_trends", map[string]any{"metric": "vo2max"}, nil); !res.IsError {
		t.Error("enum not enforced")
	}
	if res := call(t, cs, "get_trends", map[string]any{"metric": "hrv", "from": "2020-01-01"}, nil); !res.IsError {
		t.Error("range limit not enforced")
	}
}

func TestCoachNotes(t *testing.T) {
	store := &fakeStore{}
	cs := connect(t, store)

	if res := call(t, cs, "save_coach_note", map[string]any{"kind": "daily", "content": " "}, nil); !res.IsError {
		t.Error("empty note accepted")
	}
	var saved NoteOut
	call(t, cs, "save_coach_note", map[string]any{"kind": "decision", "content": "Długie skrócone do 16 km: HRV poniżej bazy."}, &saved)
	if saved.ID != 1 || saved.Date != "2026-10-25" || saved.Kind != "decision" {
		t.Errorf("saved: %+v", saved)
	}
	store.notes = append([]Note{{ID: 0, Date: d("2026-09-01"), Kind: "daily", Content: "stara"}}, store.notes...)

	var log CoachLog
	call(t, cs, "get_coach_log", nil, &log)
	if log.Since != "2026-10-12" || len(log.Notes) != 1 || log.Notes[0].Content != saved.Content {
		t.Errorf("log: %+v", log)
	}
}

func TestHandlerAuth(t *testing.T) {
	token := strings.Repeat("x", 32)
	srv := httptest.NewServer(Handler(New(&fakeStore{}, time.Now), token))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	for _, c := range []struct {
		auth string
		want int
	}{
		{"", http.StatusUnauthorized},
		{"Bearer wrong", http.StatusUnauthorized},
		{"Bearer " + token, http.StatusOK},
	} {
		req, _ := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if c.auth != "" {
			req.Header.Set("Authorization", c.auth)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != c.want {
			t.Errorf("auth %q: status %d, want %d", c.auth, res.StatusCode, c.want)
		}
	}
}
