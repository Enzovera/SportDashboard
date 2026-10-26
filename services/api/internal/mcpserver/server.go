// Package mcpserver is the MCP server: the only way AI reaches the data (docs/design.md, "Narzędzia MCP").
// It reads projections and writes only coach notes; there is deliberately no tool that accepts a plan change
// or writes to Garmin, Hevy or Fitatu. The same server backs the cron runs (`claude -p`) and the chat.
package mcpserver

import (
	"context"
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"forma/api/internal/dashboard"
)

// Version is reported to MCP clients.
const Version = "0.1.0"

type tools struct {
	store Store
	now   func() time.Time
}

func (t *tools) today() time.Time { return dashboard.Today(t.now()) }

// New registers the tools. Stage 2/4 tools (nutrition, supplements, plan, profile, propose_plan_change) are added
// once their projections exist, so the model never sees a tool that can only answer "no data".
func New(store Store, now func() time.Time) *mcp.Server {
	t := &tools{store: store, now: now}
	s := mcp.NewServer(&mcp.Implementation{Name: "forma", Version: Version}, nil)
	read := &mcp.ToolAnnotations{ReadOnlyHint: true}

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_day_summary",
		Description: "One day: sleep, HRV (night, 7-day mean, 60-day baseline), resting HR, steps, kcal burned, weight " +
			"(day and 7-day mean), training load (ATL/CTL end of day, TSB = form going into the day), activities, " +
			"strength sessions and sync status. Null means no data. `missing` lists sources not available.",
		Annotations: read,
	}, t.daySummary)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_activities",
		Description: "Garmin activities in a date range with distance, duration, pace (runs), average HR and training load.",
		Annotations: read,
	}, t.activities)

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_strength_volume",
		Description: "Hevy strength volume in a date range: working sets (no warm-ups) and tonnage, grouped by ISO week, " +
			"muscle group or session.",
		InputSchema: schemaFor[strengthIn](map[string][]any{"group_by": strengthGroups}),
		Annotations: read,
	}, t.strengthVolume)

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_trends",
		Description: "Series of one computed metric. Daily: training_load, atl, ctl, tsb, hrv, hrv_7d, resting_hr, " +
			"sleep_h, weight, weight_7d. Weekly (Monday dates): run_km_week, long_run_km_week, strength_sets_week; " +
			"each weekly point has change_pct against the week before. Null means no data for that point.",
		InputSchema: schemaFor[trendIn](map[string][]any{"metric": trendNames()}),
		Annotations: read,
	}, t.trends)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_coach_log",
		Description: "Recent coach notes (daily and weekly summaries, decisions with reasons), newest first.",
		Annotations: read,
	}, t.coachLog)

	mcp.AddTool(s, &mcp.Tool{
		Name: "save_coach_note",
		Description: "Append a note to the coach log: a daily or weekly summary, or a decision with its reason. " +
			"Notes are the coach's only memory between runs.",
		InputSchema: schemaFor[noteIn](map[string][]any{"kind": noteKinds}),
	}, t.saveNote)
	return s
}

// schemaFor infers the input schema from T and adds enums the struct tags cannot express.
func schemaFor[T any](enums map[string][]any) *jsonschema.Schema {
	s, err := jsonschema.For[T](nil)
	if err != nil {
		panic(err)
	}
	for prop, values := range enums {
		s.Properties[prop].Enum = values
	}
	return s
}

// Handler serves MCP over streamable HTTP. Stateless: every request is self-contained, so restarts and
// several clients (cron run and chat at once) need no session bookkeeping. Requests need `Bearer <token>`.
func Handler(s *mcp.Server, token string) http.Handler {
	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s },
		&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), want) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// ---------- handlers ----------

type dateIn struct {
	Date string `json:"date,omitempty" jsonschema:"YYYY-MM-DD, default today (Europe/Warsaw)"`
}

type rangeIn struct {
	From string `json:"from,omitempty" jsonschema:"YYYY-MM-DD, default 27 days before to"`
	To   string `json:"to,omitempty" jsonschema:"YYYY-MM-DD, default today; at most 366 days in total"`
}

type activitiesIn struct {
	rangeIn
	Type string `json:"type,omitempty" jsonschema:"Garmin sport filter by substring, e.g. running (also treadmill and trail), cycling, strength"`
}

type strengthIn struct {
	rangeIn
	GroupBy string `json:"group_by,omitempty" jsonschema:"week (default), muscle or session"`
}

type trendIn struct {
	rangeIn
	Metric string `json:"metric"`
}

type logIn struct {
	Days int `json:"days,omitempty" jsonschema:"how many days back, including today; default 14, max 90"`
}

type noteIn struct {
	Kind    string `json:"kind"`
	Date    string `json:"date,omitempty" jsonschema:"YYYY-MM-DD the note is about, default today"`
	Content string `json:"content" jsonschema:"the note in Polish, plain text or short markdown"`
}

// load returns projections with enough history before `from` for ATL/CTL and the 60-day HRV baseline.
func (t *tools) load(ctx context.Context, from time.Time) (dashboard.Data, error) {
	return t.store.Load(ctx, from.Add(-historyDays*day))
}

func (t *tools) daySummary(ctx context.Context, _ *mcp.CallToolRequest, in dateIn) (*mcp.CallToolResult, DaySummary, error) {
	date, err := parseDate(in.Date, t.today())
	if err != nil {
		return nil, DaySummary{}, err
	}
	data, err := t.load(ctx, date)
	if err != nil {
		return nil, DaySummary{}, internal(err)
	}
	detail, err := t.store.Day(ctx, date)
	if err != nil {
		return nil, DaySummary{}, internal(err)
	}
	return nil, BuildDaySummary(data, detail, date, t.now()), nil
}

func (t *tools) activities(ctx context.Context, _ *mcp.CallToolRequest, in activitiesIn) (*mcp.CallToolResult, ActivityList, error) {
	from, to, err := parseRange(in.From, in.To, t.today())
	if err != nil {
		return nil, ActivityList{}, err
	}
	data, err := t.store.Load(ctx, from)
	if err != nil {
		return nil, ActivityList{}, internal(err)
	}
	return nil, BuildActivities(data, from, to, in.Type), nil
}

func (t *tools) strengthVolume(ctx context.Context, _ *mcp.CallToolRequest, in strengthIn) (*mcp.CallToolResult, StrengthVolume, error) {
	from, to, err := parseRange(in.From, in.To, t.today())
	if err != nil {
		return nil, StrengthVolume{}, err
	}
	if in.GroupBy == "" {
		in.GroupBy = "week"
	}
	data, err := t.store.Load(ctx, from)
	if err != nil {
		return nil, StrengthVolume{}, internal(err)
	}
	return nil, BuildStrengthVolume(data, from, to, in.GroupBy), nil
}

func (t *tools) trends(ctx context.Context, _ *mcp.CallToolRequest, in trendIn) (*mcp.CallToolResult, Trend, error) {
	from, to, err := parseRange(in.From, in.To, t.today())
	if err != nil {
		return nil, Trend{}, err
	}
	data, err := t.load(ctx, from)
	if err != nil {
		return nil, Trend{}, internal(err)
	}
	trend, err := BuildTrend(data, in.Metric, from, to)
	return nil, trend, err
}

func (t *tools) coachLog(ctx context.Context, _ *mcp.CallToolRequest, in logIn) (*mcp.CallToolResult, CoachLog, error) {
	since, err := logSince(in.Days, t.today())
	if err != nil {
		return nil, CoachLog{}, err
	}
	notes, err := t.store.Notes(ctx, since)
	if err != nil {
		return nil, CoachLog{}, internal(err)
	}
	out := CoachLog{Since: iso(since), Notes: []NoteOut{}}
	for _, n := range notes {
		out.Notes = append(out.Notes, toNote(n))
	}
	return nil, out, nil
}

func (t *tools) saveNote(ctx context.Context, _ *mcp.CallToolRequest, in noteIn) (*mcp.CallToolResult, NoteOut, error) {
	date, content, err := validateNote(in.Kind, in.Date, in.Content, t.today())
	if err != nil {
		return nil, NoteOut{}, err
	}
	n, err := t.store.SaveNote(ctx, date, in.Kind, content)
	if err != nil {
		return nil, NoteOut{}, internal(err)
	}
	slog.Info("coach note saved", "id", n.ID, "date", iso(n.Date), "kind", n.Kind)
	return nil, toNote(n), nil
}

// internal logs a storage error and hides its details from the model.
func internal(err error) error {
	slog.Error("mcp tool failed", "err", err)
	return errInternal
}

var errInternal = errors.New("internal error, try again later")
