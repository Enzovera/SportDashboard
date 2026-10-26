// Package plan is the event-sourced training plan (docs/design.md, "Stan trenera"): event payloads, the fold
// into the current state, validators, and the store that appends events and rewrites the projections.
// The coach only proposes (Propose, called by the MCP server); only a human decides (Decide, dashboard API).
package plan

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event types, past tense, payload versioned with "v".
const (
	PlanCreated        = "PlanCreated"
	PlanChangeProposed = "PlanChangeProposed"
	PlanChangeAccepted = "PlanChangeAccepted"
	PlanChangeRejected = "PlanChangeRejected"
)

const PayloadVersion = 1

// Event is one plan_event row.
type Event struct {
	StreamID  string
	Seq       int
	Type      string
	Payload   json.RawMessage
	Actor     string // user | coach
	CreatedAt time.Time
}

// Date is a calendar day serialized as YYYY-MM-DD.
type Date struct{ time.Time }

func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d.Format(time.DateOnly)) }

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return fmt.Errorf("date %q: %w", s, err)
	}
	d.Time = t
	return nil
}

func NewDate(t time.Time) Date {
	return Date{time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

// ---------- PlanCreated ----------

type Created struct {
	V       int      `json:"v"`
	Goal    Goal     `json:"goal"`
	Blocks  []Block  `json:"blocks"`
	Targets []Target `json:"targets"` // per day type, from the first day of the plan
	Days    []Day    `json:"days"`    // days not listed are rest days
}

type Goal struct {
	Race             string      `json:"race"`
	Date             Date        `json:"date"`
	TargetTimeS      *int        `json:"targetTimeS"`
	TargetPaceSPerKm *int        `json:"targetPaceSPerKm"`
	Secondary        []Secondary `json:"secondary"`
}

type Secondary struct {
	Label  string  `json:"label"`
	Value  string  `json:"value"`
	Detail *string `json:"detail"`
}

type Block struct {
	Kind    string  `json:"kind"` // base | build | peak | taper
	From    Date    `json:"from"`
	To      Date    `json:"to"`
	Summary *string `json:"summary"`
}

// Target is the nutrition target of a day type. In a change, nil fields keep the current value.
type Target struct {
	DayType  string   `json:"dayType"`
	Kcal     *float64 `json:"kcal"`
	ProteinG *float64 `json:"proteinG"`
	CarbsG   *float64 `json:"carbsG"`
	FatG     *float64 `json:"fatG"`
}

type Day struct {
	Date     Date      `json:"date"`
	DayType  string    `json:"dayType"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
	Kind       string   `json:"kind"`  // run | strength
	Title      string   `json:"title"` // e.g. "Długie wybieganie", "Push", "bieg łatwy"
	Detail     *string  `json:"detail"`
	DistanceKm *float64 `json:"distanceKm"`
	Short      *string  `json:"short"` // label for the week diff, e.g. "Długie 16 km"; derived when nil
}

// ---------- PlanChangeProposed ----------

// Proposed is written by the coach (MCP propose_plan_change) after the validators passed.
type Proposed struct {
	V             int     `json:"v"`
	ProposalID    string  `json:"proposalId"`
	Kind          string  `json:"kind"`   // nutrition | strength | run
	Origin        string  `json:"origin"` // daily_run | weekly_run | chat
	Title         string  `json:"title"`
	Summary       string  `json:"summary"` // short, per affected day, e.g. "12 serii"
	Rationale     string  `json:"rationale"`
	EffectiveFrom *Date   `json:"effectiveFrom"`
	Diff          Diff    `json:"diff"`
	Change        Change  `json:"change"`
	Checks        []Check `json:"checks"`
}

// Change is what accepting a proposal applies.
type Change struct {
	Days     []Day     `json:"days,omitempty"`    // replace these days wholesale (no sessions = rest day)
	Targets  []Target  `json:"targets,omitempty"` // from EffectiveFrom (else the decision day)
	Reminder *Reminder `json:"reminder,omitempty"`
}

// Reminder moves a supplement reminder time; the coach can change nothing else about supplements.
type Reminder struct {
	SupplementID   string `json:"supplementId"`
	SupplementName string `json:"supplementName"`
	From           string `json:"from"` // "HH:MM"
	To             string `json:"to"`
}

// Diff is the before/after table written by the coach for the Proposals view.
type Diff struct {
	Heading string    `json:"heading"`
	Rows    []DiffRow `json:"rows"`
	Total   *DiffRow  `json:"total"`
}

type DiffRow struct {
	Label  string  `json:"label"`
	Before any     `json:"before"` // number | string | null
	After  any     `json:"after"`
	Unit   *string `json:"unit"`
	Note   *string `json:"note"`
}

// Check is one validator result.
type Check struct {
	Rule   string `json:"rule"`
	Status string `json:"status"` // pass | warn | fail
	Text   string `json:"text"`
}

// ---------- PlanChangeAccepted / PlanChangeRejected ----------

type Decided struct {
	V          int     `json:"v"`
	ProposalID string  `json:"proposalId"`
	Comment    *string `json:"comment"`
}
