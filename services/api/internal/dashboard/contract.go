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

// ---------- GET /api/body ----------

type BodyResponse struct {
	Date           string          `json:"date"`
	From           string          `json:"from"`
	To             string          `json:"to"`
	RangeWeeks     int             `json:"rangeWeeks"`
	Device         *string         `json:"device"`
	Sync           []SyncStatus    `json:"sync"`
	Composition    BodyComposition `json:"composition"`
	TrendKgPerWeek *float64        `json:"trendKgPerWeek"`
	Weeks          []BodyWeek      `json:"weeks"`
	WeighIns       []WeighIn       `json:"weighIns"`
	Goal           any             `json:"goal"`      // stage 4 (athlete profile)
	CoachNote      any             `json:"coachNote"` // stage 3
}

type BodyComposition struct {
	MeasuredAt    *string  `json:"measuredAt"`
	LastKg        *float64 `json:"lastKg"`
	Avg7dKg       *float64 `json:"avg7dKg"`
	Avg7dChangeKg *float64 `json:"avg7dChangeKg"`
	FatPct        *float64 `json:"fatPct"`
	MuscleKg      *float64 `json:"muscleKg"`
	MuscleTrend   *string  `json:"muscleTrend"` // up | stable | down
	WeekCount     int      `json:"weekCount"`
	WeekDays      int      `json:"weekDays"`
	WeekMissing   []string `json:"weekMissing"`
	RangeCount    int      `json:"rangeCount"`
}

type BodyWeek struct {
	WeekStart string   `json:"weekStart"`
	AvgKg     *float64 `json:"avgKg"`
	MinKg     *float64 `json:"minKg"`
	MaxKg     *float64 `json:"maxKg"`
	Count     int      `json:"count"`
}

type WeighIn struct {
	MeasuredAt string   `json:"measuredAt"`
	WeightKg   float64  `json:"weightKg"`
	FatPct     *float64 `json:"fatPct"`
	MuscleKg   *float64 `json:"muscleKg"`
	WaterPct   *float64 `json:"waterPct"`
}

// ---------- GET /api/nutrition?range=14d ----------

type NutritionRow struct {
	Eaten  float64 `json:"eaten"`
	Target float64 `json:"target"`
}

type NutritionResponse struct {
	Range           string                `json:"range"`
	From            string                `json:"from"`
	To              string                `json:"to"`
	Sync            []SyncStatus          `json:"sync"`
	Today           *NutritionTodayDetail `json:"today"`
	Balance         *NutritionBalance     `json:"balance"`
	Days            []NutritionDay        `json:"days"`
	Targets         []DayTarget           `json:"targets"`
	TargetProposals []TargetProposal      `json:"targetProposals"`
	Meals           []Meal                `json:"meals"`
	CoachNote       any                   `json:"coachNote"` // stage 3
	Supplements     *SupplementsToday     `json:"supplements"`
}

type NutritionTodayDetail struct {
	DayType  string       `json:"dayType"`
	Kcal     NutritionRow `json:"kcal"`
	ProteinG NutritionRow `json:"proteinG"`
	CarbsG   NutritionRow `json:"carbsG"`
	FatG     NutritionRow `json:"fatG"`
}

type NutritionBalance struct {
	From                string        `json:"from"`
	To                  string        `json:"to"`
	KcalAvg             float64       `json:"kcalAvg"`
	KcalTargetAvg       float64       `json:"kcalTargetAvg"`
	KcalVsTargetPct     float64       `json:"kcalVsTargetPct"`
	ProteinAvgGPerKg    float64       `json:"proteinAvgGPerKg"`
	ProteinTargetGPerKg float64       `json:"proteinTargetGPerKg"`
	ProteinMinGPerKg    float64       `json:"proteinMinGPerKg"`
	ProteinDaysAtMin    int           `json:"proteinDaysAtMin"`
	Days                int           `json:"days"`
	LongRunCarbs        *LongRunCarbs `json:"longRunCarbs"`
}

type LongRunCarbs struct {
	Date         string  `json:"date"`
	GPerKg       float64 `json:"gPerKg"`
	TargetGPerKg float64 `json:"targetGPerKg"`
	Tone         string  `json:"tone"`
}

type NutritionDay struct {
	Date            string   `json:"date"`
	DayType         *string  `json:"dayType"`
	Complete        bool     `json:"complete"`
	Kcal            *float64 `json:"kcal"`
	KcalTarget      *float64 `json:"kcalTarget"`
	KcalVsTargetPct *float64 `json:"kcalVsTargetPct"`
	KcalTone        string   `json:"kcalTone"`
	ProteinGPerKg   *float64 `json:"proteinGPerKg"`
	ProteinTone     string   `json:"proteinTone"`
}

type DayTarget struct {
	DayType  string  `json:"dayType"`
	Kcal     float64 `json:"kcal"`
	ProteinG float64 `json:"proteinG"`
	CarbsG   float64 `json:"carbsG"`
	FatG     float64 `json:"fatG"`
}

type TargetProposal struct {
	ID      string  `json:"id"`
	DayType string  `json:"dayType"`
	Field   string  `json:"field"` // kcal | proteinG | carbsG | fatG
	Value   float64 `json:"value"`
}

type Meal struct {
	ID       string  `json:"id"`
	Slot     string  `json:"slot"`
	At       *string `json:"at"`
	Title    string  `json:"title"`
	Kcal     float64 `json:"kcal"`
	ProteinG float64 `json:"proteinG"`
	CarbsG   float64 `json:"carbsG"`
	FatG     float64 `json:"fatG"`
}

type SupplementsToday struct {
	Taken int     `json:"taken"`
	Total int     `json:"total"`
	Next  *string `json:"next"`
}
