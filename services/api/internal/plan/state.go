package plan

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"forma/api/internal/projector"
)

// Day types (nutrition target type of a day).
const (
	DayRest        = "rest"
	DayStrength    = "strength"
	DayRun         = "run"
	DayStrengthRun = "strength_run"
	DayLongRun     = "long_run"
)

var DayTypes = []string{DayRest, DayStrength, DayRun, DayStrengthRun, DayLongRun}

// DayTarget is a complete nutrition target of a day type, in force from a date.
type DayTarget struct {
	From     time.Time
	DayType  string
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
}

// Proposal is a PlanChangeProposed with its decision, if any.
type Proposal struct {
	Proposed
	ProposedAt time.Time
	Decision   *Decision
}

type Decision struct {
	Accepted bool
	At       time.Time
	Comment  *string
}

// State is the plan folded from the events of the current stream.
type State struct {
	StreamID  string
	CreatedAt time.Time
	Goal      Goal
	Blocks    []Block
	Days      map[time.Time]Day // only days with an entry; others are rest days
	Targets   []DayTarget       // ordered by From
	Proposals []Proposal        // in event order
	LastSeq   int
}

// Fold applies events (ordered by seq) of one stream. nil state for no events.
func Fold(events []Event) (*State, error) {
	var s *State
	for _, e := range events {
		if s == nil && e.Type != PlanCreated {
			return nil, fmt.Errorf("plan_event %s/%d: %s before PlanCreated", e.StreamID, e.Seq, e.Type)
		}
		switch e.Type {
		case PlanCreated:
			var p Created
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				return nil, fmt.Errorf("PlanCreated: %w", err)
			}
			s = &State{StreamID: e.StreamID, CreatedAt: e.CreatedAt, Goal: p.Goal, Blocks: p.Blocks, Days: map[time.Time]Day{}}
			for _, d := range p.Days {
				s.Days[d.Date.Time] = d
			}
			from := s.firstDay()
			for _, t := range p.Targets {
				s.Targets = append(s.Targets, DayTarget{From: from, DayType: t.DayType, Kcal: val(t.Kcal),
					ProteinG: val(t.ProteinG), CarbsG: val(t.CarbsG), FatG: val(t.FatG)})
			}
		case PlanChangeProposed:
			var p Proposed
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				return nil, fmt.Errorf("PlanChangeProposed: %w", err)
			}
			s.Proposals = append(s.Proposals, Proposal{Proposed: p, ProposedAt: e.CreatedAt})
		case PlanChangeAccepted, PlanChangeRejected:
			var p Decided
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				return nil, fmt.Errorf("%s: %w", e.Type, err)
			}
			pr := s.Proposal(p.ProposalID)
			if pr == nil || pr.Decision != nil {
				return nil, fmt.Errorf("%s: proposal %q unknown or already decided", e.Type, p.ProposalID)
			}
			pr.Decision = &Decision{Accepted: e.Type == PlanChangeAccepted, At: e.CreatedAt, Comment: p.Comment}
			if pr.Decision.Accepted {
				s.apply(pr.Change, pr.effectiveFrom(e.CreatedAt))
			}
		default:
			return nil, fmt.Errorf("plan_event %s/%d: unknown type %q", e.StreamID, e.Seq, e.Type)
		}
		s.LastSeq = e.Seq
	}
	return s, nil
}

// effectiveFrom is the first day a change applies: as proposed, else the local day of the decision.
func (p *Proposal) effectiveFrom(decidedAt time.Time) time.Time {
	if p.EffectiveFrom != nil {
		return p.EffectiveFrom.Time
	}
	return NewDate(decidedAt.In(projector.Warsaw)).Time
}

func (s *State) apply(c Change, from time.Time) {
	for _, d := range c.Days {
		s.Days[d.Date.Time] = d
	}
	for _, t := range c.Targets {
		cur := s.TargetAt(from, t.DayType)
		next := DayTarget{From: from, DayType: t.DayType}
		if cur != nil {
			next.Kcal, next.ProteinG, next.CarbsG, next.FatG = cur.Kcal, cur.ProteinG, cur.CarbsG, cur.FatG
		}
		if t.Kcal != nil {
			next.Kcal = *t.Kcal
		}
		if t.ProteinG != nil {
			next.ProteinG = *t.ProteinG
		}
		if t.CarbsG != nil {
			next.CarbsG = *t.CarbsG
		}
		if t.FatG != nil {
			next.FatG = *t.FatG
		}
		// Replace a version from the same day, otherwise add one.
		kept := s.Targets[:0:0]
		for _, x := range s.Targets {
			if !(x.DayType == t.DayType && x.From.Equal(from)) {
				kept = append(kept, x)
			}
		}
		s.Targets = append(kept, next)
		sort.SliceStable(s.Targets, func(i, j int) bool { return s.Targets[i].From.Before(s.Targets[j].From) })
	}
}

// With returns a copy of the state with a change applied from `from` (preview and validation).
func (s *State) With(c Change, from time.Time) *State {
	cp := *s
	cp.Days = make(map[time.Time]Day, len(s.Days))
	for k, v := range s.Days {
		cp.Days[k] = v
	}
	cp.Targets = append([]DayTarget{}, s.Targets...)
	cp.apply(c, from)
	return &cp
}

// Day returns the planned day; a day without an entry is a rest day.
func (s *State) Day(date time.Time) Day {
	if d, ok := s.Days[date]; ok {
		return d
	}
	return Day{Date: Date{date}, DayType: DayRest}
}

// TargetAt is the target of a day type in force on a date; nil when none is set.
func (s *State) TargetAt(date time.Time, dayType string) *DayTarget {
	var out *DayTarget
	for i := range s.Targets {
		t := &s.Targets[i]
		if t.DayType == dayType && !t.From.After(date) {
			out = t
		}
	}
	if out == nil { // before the first version: the earliest one
		for i := range s.Targets {
			if s.Targets[i].DayType == dayType {
				return &s.Targets[i]
			}
		}
	}
	return out
}

func (s *State) Proposal(id string) *Proposal {
	for i := range s.Proposals {
		if s.Proposals[i].ProposalID == id {
			return &s.Proposals[i]
		}
	}
	return nil
}

// Pending returns undecided proposals, newest first.
func (s *State) Pending() []Proposal {
	var out []Proposal
	for i := len(s.Proposals) - 1; i >= 0; i-- {
		if s.Proposals[i].Decision == nil {
			out = append(out, s.Proposals[i])
		}
	}
	return out
}

// BlockAt returns the block containing date and the 1-based week number in it.
func (s *State) BlockAt(date time.Time) (*Block, int) {
	for i := range s.Blocks {
		b := &s.Blocks[i]
		if !date.Before(b.From.Time) && !date.After(b.To.Time) {
			return b, int(date.Sub(b.From.Time).Hours()/24)/7 + 1
		}
	}
	return nil, 0
}

// RunKm is the planned running distance of a day; 0 for none.
func (d Day) RunKm() float64 {
	var km float64
	for _, x := range d.Sessions {
		if x.Kind == "run" && x.DistanceKm != nil {
			km += *x.DistanceKm
		}
	}
	return km
}

func (s *State) firstDay() time.Time {
	var first time.Time
	for d := range s.Days {
		if first.IsZero() || d.Before(first) {
			first = d
		}
	}
	for _, b := range s.Blocks {
		if first.IsZero() || b.From.Before(first) {
			first = b.From.Time
		}
	}
	return first
}

func val(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// Covers reports whether the plan says what a date is: from its first day to race day.
func (s *State) Covers(date time.Time) bool {
	return !date.Before(s.firstDay()) && !date.After(s.Goal.Date.Time)
}
