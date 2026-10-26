package dashboard

import "forma/api/internal/plan"

// JSON contract of the Plan and Proposals views. Mirrors web/src/api/plan.ts and web/src/api/proposals.ts.

// ---------- GET /api/plan ----------

type PlanResponse struct {
	Date             string    `json:"date"`
	Plan             *PlanView `json:"plan"`
	ProposalsPending *int      `json:"proposalsPending"`
}

type PlanView struct {
	Goal        RaceGoal    `json:"goal"`
	Blocks      []PlanBlock `json:"blocks"`
	CurrentWeek *PlanWeek   `json:"currentWeek"`
	NextWeek    *PlanWeek   `json:"nextWeek"`
}

type RaceGoal struct {
	Race             string           `json:"race"`
	Date             string           `json:"date"`
	WeeksLeft        int              `json:"weeksLeft"`
	TargetTimeS      *int             `json:"targetTimeS"`
	TargetPaceSPerKm *int             `json:"targetPaceSPerKm"`
	Secondary        []plan.Secondary `json:"secondary"`
}

type PlanBlock struct {
	Kind    string  `json:"kind"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Weeks   int     `json:"weeks"`
	Summary *string `json:"summary"`
}

type PlanWeek struct {
	From            string    `json:"from"`
	To              string    `json:"to"`
	Block           *string   `json:"block"`
	WeekInBlock     *int      `json:"weekInBlock"`
	PlannedKm       float64   `json:"plannedKm"`
	PlannedStrength int       `json:"plannedStrength"`
	DoneKm          *float64  `json:"doneKm"`
	DoneStrength    *int      `json:"doneStrength"`
	Days            []PlanDay `json:"days"`
}

type PlanDay struct {
	Date      string        `json:"date"`
	Sessions  []PlanSession `json:"sessions"`
	RunKm     *float64      `json:"runKm"`
	DayType   string        `json:"dayType"`
	Target    *Macros       `json:"target"`
	Done      *DayDone      `json:"done"`
	Proposals []DayProposal `json:"proposals"`
}

type PlanSession struct {
	Kind       string   `json:"kind"`
	Title      string   `json:"title"`
	Detail     *string  `json:"detail"`
	DistanceKm *float64 `json:"distanceKm"`
}

type Macros struct {
	Kcal     float64 `json:"kcal"`
	ProteinG float64 `json:"proteinG"`
	CarbsG   float64 `json:"carbsG"`
	FatG     float64 `json:"fatG"`
}

type DayDone struct {
	RunKm        *float64 `json:"runKm"`
	StrengthSets *int     `json:"strengthSets"`
}

type DayProposal struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
}

// ---------- GET /api/proposals, POST /api/proposals/{id}/accept|reject ----------

type ProposalsResponse struct {
	Date    string             `json:"date"`
	Pending []ProposalView     `json:"pending"`
	Blocked []BlockedView      `json:"blocked"`
	History []ProposalDecision `json:"history"`
	Week    *WeekDiff          `json:"week"`
}

type ProposalView struct {
	ID            string       `json:"id"`
	Kind          string       `json:"kind"`
	Origin        string       `json:"origin"`
	ProposedAt    string       `json:"proposedAt"`
	Title         string       `json:"title"`
	Rationale     string       `json:"rationale"`
	EffectiveFrom *string      `json:"effectiveFrom"`
	Diff          plan.Diff    `json:"diff"`
	Checks        []plan.Check `json:"checks"`
}

type BlockedView struct {
	ID         string       `json:"id"`
	Kind       string       `json:"kind"`
	Origin     string       `json:"origin"`
	ProposedAt string       `json:"proposedAt"`
	Title      string       `json:"title"`
	Checks     []plan.Check `json:"checks"`
}

type ProposalDecision struct {
	ID         string  `json:"id"`
	Kind       string  `json:"kind"`
	Origin     string  `json:"origin"`
	Title      string  `json:"title"`
	ProposedAt string  `json:"proposedAt"`
	DecidedAt  string  `json:"decidedAt"`
	Decision   string  `json:"decision"` // accepted | rejected
	Comment    *string `json:"comment"`
}

type WeekDiff struct {
	From  string        `json:"from"`
	To    string        `json:"to"`
	Block *string       `json:"block"`
	Days  []WeekDiffDay `json:"days"`
}

type WeekDiffDay struct {
	Date    string      `json:"date"`
	Now     DayPlan     `json:"now"`
	Changes []DayChange `json:"changes"`
}

type DayChange struct {
	ProposalID string  `json:"proposalId"`
	After      DayPlan `json:"after"`
}

type DayPlan struct {
	Session string   `json:"session"`
	RunKm   *float64 `json:"runKm"`
	Kcal    float64  `json:"kcal"`
	CarbsG  float64  `json:"carbsG"`
}
