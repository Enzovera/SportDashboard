package dashboard

// JSON contract of the dashboard API. Mirrors web/src/api/types.ts field by field: change both together.
// Nullable fields are pointers (or `any` set to nil) so they serialize as null, never omitted.

type SyncStatus struct {
	Source        string  `json:"source"`
	Status        string  `json:"status"` // ok | partial | error | stale
	LastSuccessAt *string `json:"lastSuccessAt"`
}

// ---------- GET /api/today ----------

type TodayResponse struct {
	Date             string       `json:"date"`
	RaceWeeksLeft    *int         `json:"raceWeeksLeft"`
	Sync             []SyncStatus `json:"sync"`
	Coach            any          `json:"coach"`   // stage 3
	Session          any          `json:"session"` // stage 4
	Week             WeekGrid     `json:"week"`
	Tomorrow         any          `json:"tomorrow"` // stage 4
	Readiness        Readiness    `json:"readiness"`
	Nutrition        any          `json:"nutrition"`   // stage 2
	Supplements      any          `json:"supplements"` // stage 2
	ProposalsPending *int         `json:"proposalsPending"`
	Weight           *WeightTrend `json:"weight"`
}

type WeekGrid struct {
	From            string    `json:"from"`
	To              string    `json:"to"`
	Days            []WeekDay `json:"days"`
	RunKmDone       float64   `json:"runKmDone"`
	RunKmPlanned    *float64  `json:"runKmPlanned"`
	StrengthDone    int       `json:"strengthDone"`
	StrengthPlanned *int      `json:"strengthPlanned"`
}

type WeekDay struct {
	Date       string   `json:"date"`
	RunKm      *float64 `json:"runKm"`
	RunPlanned bool     `json:"runPlanned"`
	Strength   *string  `json:"strength"`
}

type Readiness struct {
	MeasuredAt       *string  `json:"measuredAt"`
	HRV7dMs          *float64 `json:"hrv7dMs"`
	HRVBaselineMs    *float64 `json:"hrvBaselineMs"`
	HRVTone          string   `json:"hrvTone"` // good | warn | neutral
	HRVLabel         *string  `json:"hrvLabel"`
	SleepS           *int     `json:"sleepS"`
	SleepAvg7dS      *int     `json:"sleepAvg7dS"`
	RestingHR        *int     `json:"restingHr"`
	RestingHRDelta7d *int     `json:"restingHrDelta7d"`
	TSB              *float64 `json:"tsb"`
	Form             *string  `json:"form"`
}

type WeightTrend struct {
	Avg7dKg        float64  `json:"avg7dKg"`
	DeltaPerWeekKg *float64 `json:"deltaPerWeekKg"`
}

// ---------- GET /api/training?range=4w|12w|block ----------

type TrainingResponse struct {
	Range            string        `json:"range"`
	From             string        `json:"from"`
	To               string        `json:"to"`
	Sync             []SyncStatus  `json:"sync"`
	Load             LoadSummary   `json:"load"`
	WeeklyKm         []WeeklyKm    `json:"weeklyKm"`
	TodayPlannedKm   *float64      `json:"todayPlannedKm"`
	LoadSeries       []LoadPoint   `json:"loadSeries"`
	CoachNote        any           `json:"coachNote"` // stage 3
	RecentRuns       []RunRow      `json:"recentRuns"`
	SetsPerMuscle    SetsPerMuscle `json:"setsPerMuscle"`
	StrengthSessions []StrengthRow `json:"strengthSessions"`
}

type LoadSummary struct {
	WeekFrom      string   `json:"weekFrom"`
	WeekTo        string   `json:"weekTo"`
	WeekKm        float64  `json:"weekKm"`
	PrevWeekKm    *float64 `json:"prevWeekKm"`
	CTL           *float64 `json:"ctl"`
	CTLDeltaRange *float64 `json:"ctlDeltaRange"`
	ATL           *float64 `json:"atl"`
	ATLDeltaWeek  *float64 `json:"atlDeltaWeek"`
	TSB           *float64 `json:"tsb"`
	Form          *string  `json:"form"`
}

type WeeklyKm struct {
	WeekStart string   `json:"weekStart"`
	Km        float64  `json:"km"`
	PlannedKm *float64 `json:"plannedKm"`
	Deload    bool     `json:"deload"`
}

type LoadPoint struct {
	Date string  `json:"date"`
	CTL  float64 `json:"ctl"`
	ATL  float64 `json:"atl"`
}

type RunRow struct {
	ID         string   `json:"id"`
	Date       string   `json:"date"`
	Title      string   `json:"title"`
	DistanceKm float64  `json:"distanceKm"`
	PaceSPerKm *float64 `json:"paceSPerKm"`
	AvgHR      *int     `json:"avgHr"`
	Load       *float64 `json:"load"`
}

type SetsPerMuscle struct {
	WeekFrom string       `json:"weekFrom"`
	WeekTo   string       `json:"weekTo"`
	Items    []MuscleSets `json:"items"`
}

type MuscleSets struct {
	Muscle string `json:"muscle"`
	Sets   int    `json:"sets"`
}

type StrengthRow struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"`
	Title       string  `json:"title"`
	DurationMin int     `json:"durationMin"`
	Sets        int     `json:"sets"`
	TonnageKg   float64 `json:"tonnageKg"`
}
