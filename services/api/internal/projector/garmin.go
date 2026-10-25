package projector

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

// Day is one day_summary row, merged from several Garmin endpoints for the same calendar date.
type Day struct {
	Date           time.Time
	SleepS         *int
	SleepScore     *int
	HRVLastNightMs *int
	HRVWeeklyAvgMs *int
	HRVStatus      *string
	RestingHR      *int
	Steps          *int
	ActiveKcal     *int
	TotalKcal      *int
	WeightKg       *float64
}

// Activity is one activity row (Garmin).
type Activity struct {
	ExternalID     string
	StartedAt      time.Time
	LocalDate      time.Time
	Sport          string
	Name           *string
	DistanceM      *float64
	DurationS      *float64
	MovingS        *float64
	AvgHR          *int
	MaxHR          *int
	ElevationGainM *float64
	TrainingLoad   *float64
	HRZonesS       []float64
}

// IsRun reports whether the sport counts towards running mileage (road, treadmill, trail, track...).
func (a Activity) IsRun() bool { return strings.Contains(a.Sport, "running") }

func ptr[T any](v T) *T { return &v }

func roundInt(f *float64) *int {
	if f == nil {
		return nil
	}
	return ptr(int(math.Round(*f)))
}

type sleepBody struct {
	DailySleepDTO struct {
		SleepTimeSeconds *float64 `json:"sleepTimeSeconds"`
		SleepScores      struct {
			Overall struct {
				Value *float64 `json:"value"`
			} `json:"overall"`
		} `json:"sleepScores"`
	} `json:"dailySleepDTO"`
}

// ApplySleep fills sleep fields from a Garmin sleep payload. Days without a recorded sleep stay nil.
func ApplySleep(d *Day, raw []byte) error {
	var b sleepBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return fmt.Errorf("sleep: %w", err)
	}
	d.SleepS = roundInt(b.DailySleepDTO.SleepTimeSeconds)
	d.SleepScore = roundInt(b.DailySleepDTO.SleepScores.Overall.Value)
	return nil
}

type hrvBody struct {
	HRVSummary struct {
		LastNightAvg *float64 `json:"lastNightAvg"`
		WeeklyAvg    *float64 `json:"weeklyAvg"`
		Status       *string  `json:"status"`
	} `json:"hrvSummary"`
}

func ApplyHRV(d *Day, raw []byte) error {
	var b hrvBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return fmt.Errorf("hrv: %w", err)
	}
	d.HRVLastNightMs = roundInt(b.HRVSummary.LastNightAvg)
	d.HRVWeeklyAvgMs = roundInt(b.HRVSummary.WeeklyAvg)
	d.HRVStatus = b.HRVSummary.Status
	return nil
}

type userSummaryBody struct {
	TotalSteps         *float64 `json:"totalSteps"`
	RestingHeartRate   *float64 `json:"restingHeartRate"`
	ActiveKilocalories *float64 `json:"activeKilocalories"`
	TotalKilocalories  *float64 `json:"totalKilocalories"`
}

func ApplyUserSummary(d *Day, raw []byte) error {
	var b userSummaryBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return fmt.Errorf("user_summary: %w", err)
	}
	d.Steps = roundInt(b.TotalSteps)
	d.RestingHR = roundInt(b.RestingHeartRate)
	d.ActiveKcal = roundInt(b.ActiveKilocalories)
	d.TotalKcal = roundInt(b.TotalKilocalories)
	return nil
}

type bodyCompositionBody struct {
	TotalAverage struct {
		Weight *float64 `json:"weight"` // grams
	} `json:"totalAverage"`
}

// ApplyBodyComposition takes the day's average weight (Xiaomi scale synced to Garmin), in kg.
func ApplyBodyComposition(d *Day, raw []byte) error {
	var b bodyCompositionBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return fmt.Errorf("body_composition: %w", err)
	}
	if w := b.TotalAverage.Weight; w != nil && *w > 0 {
		d.WeightKg = ptr(math.Round(*w/10) / 100)
	}
	return nil
}

type activityBody struct {
	ActivityID   json.Number `json:"activityId"`
	ActivityName *string     `json:"activityName"`
	ActivityType struct {
		TypeKey string `json:"typeKey"`
	} `json:"activityType"`
	StartTimeLocal       string   `json:"startTimeLocal"`
	StartTimeGMT         string   `json:"startTimeGMT"`
	Distance             *float64 `json:"distance"`
	Duration             *float64 `json:"duration"`
	MovingDuration       *float64 `json:"movingDuration"`
	AverageHR            *float64 `json:"averageHR"`
	MaxHR                *float64 `json:"maxHR"`
	ElevationGain        *float64 `json:"elevationGain"`
	ActivityTrainingLoad *float64 `json:"activityTrainingLoad"`
}

const garminTime = "2006-01-02 15:04:05"

func ParseActivity(raw []byte) (Activity, error) {
	var b activityBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return Activity{}, fmt.Errorf("activity: %w", err)
	}
	started, err := time.ParseInLocation(garminTime, b.StartTimeGMT, time.UTC)
	if err != nil {
		return Activity{}, fmt.Errorf("activity %s startTimeGMT: %w", b.ActivityID, err)
	}
	local, err := time.Parse(garminTime, b.StartTimeLocal)
	if err != nil {
		return Activity{}, fmt.Errorf("activity %s startTimeLocal: %w", b.ActivityID, err)
	}
	return Activity{
		ExternalID:     b.ActivityID.String(),
		StartedAt:      started,
		LocalDate:      time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC),
		Sport:          b.ActivityType.TypeKey,
		Name:           b.ActivityName,
		DistanceM:      b.Distance,
		DurationS:      b.Duration,
		MovingS:        b.MovingDuration,
		AvgHR:          roundInt(b.AverageHR),
		MaxHR:          roundInt(b.MaxHR),
		ElevationGainM: b.ElevationGain,
		TrainingLoad:   b.ActivityTrainingLoad,
	}, nil
}

type zone struct {
	SecsInZone float64 `json:"secsInZone"`
	ZoneNumber int     `json:"zoneNumber"`
}

// ParseZones returns seconds per HR zone, index 0 = zone 1.
func ParseZones(raw []byte) ([]float64, error) {
	var zs []zone
	if err := json.Unmarshal(raw, &zs); err != nil {
		return nil, fmt.Errorf("hr zones: %w", err)
	}
	out := make([]float64, 5)
	for _, z := range zs {
		if z.ZoneNumber >= 1 && z.ZoneNumber <= len(out) {
			out[z.ZoneNumber-1] = math.Round(z.SecsInZone)
		}
	}
	return out, nil
}
