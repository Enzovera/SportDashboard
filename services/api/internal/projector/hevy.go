package projector

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// Warsaw is the athlete's local zone: day boundaries for workouts.
var Warsaw = mustLoad("Europe/Warsaw")

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

type Workout struct {
	ID        string
	Title     *string
	StartedAt time.Time
	EndedAt   *time.Time
	LocalDate time.Time
	VolumeKg  float64
	Sets      []StrengthSet
}

type StrengthSet struct {
	ExerciseIndex      int
	SetNo              int
	ExerciseTemplateID *string
	Exercise           string
	MuscleGroup        *string
	SetType            string
	WeightKg           *float64
	Reps               *int
	RPE                *float64
	DistanceM          *float64
	DurationS          *int
}

// Working reports whether the set counts as a working set (not a warm-up, with reps done).
func (s StrengthSet) Working() bool {
	return s.SetType != "warmup" && s.Reps != nil && *s.Reps > 0
}

type workoutBody struct {
	ID        string  `json:"id"`
	Title     *string `json:"title"`
	StartTime string  `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Exercises []struct {
		Index              int     `json:"index"`
		Title              string  `json:"title"`
		ExerciseTemplateID *string `json:"exercise_template_id"`
		Sets               []struct {
			Index           int      `json:"index"`
			Type            string   `json:"type"`
			WeightKg        *float64 `json:"weight_kg"`
			Reps            *int     `json:"reps"`
			RPE             *float64 `json:"rpe"`
			DistanceMeters  *float64 `json:"distance_meters"`
			DurationSeconds *int     `json:"duration_seconds"`
		} `json:"sets"`
	} `json:"exercises"`
}

type templateBody struct {
	ID                 string  `json:"id"`
	PrimaryMuscleGroup *string `json:"primary_muscle_group"`
}

// ParseTemplate returns exercise template id -> primary muscle group.
func ParseTemplate(raw []byte) (string, *string, error) {
	var t templateBody
	if err := json.Unmarshal(raw, &t); err != nil {
		return "", nil, fmt.Errorf("exercise template: %w", err)
	}
	return t.ID, t.PrimaryMuscleGroup, nil
}

// ParseWorkout maps a Hevy workout to rows; muscles maps template ids to primary muscle groups.
func ParseWorkout(raw []byte, muscles map[string]*string) (Workout, error) {
	var b workoutBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return Workout{}, fmt.Errorf("workout: %w", err)
	}
	started, err := time.Parse(time.RFC3339, b.StartTime)
	if err != nil {
		return Workout{}, fmt.Errorf("workout %s start_time: %w", b.ID, err)
	}
	w := Workout{ID: b.ID, Title: b.Title, StartedAt: started}
	if b.EndTime != nil {
		if end, err := time.Parse(time.RFC3339, *b.EndTime); err == nil {
			w.EndedAt = &end
		}
	}
	l := started.In(Warsaw)
	w.LocalDate = time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)

	for _, ex := range b.Exercises {
		var muscle *string
		if ex.ExerciseTemplateID != nil {
			muscle = muscles[*ex.ExerciseTemplateID]
		}
		for _, s := range ex.Sets {
			set := StrengthSet{
				ExerciseIndex:      ex.Index,
				SetNo:              s.Index,
				ExerciseTemplateID: ex.ExerciseTemplateID,
				Exercise:           ex.Title,
				MuscleGroup:        muscle,
				SetType:            s.Type,
				WeightKg:           s.WeightKg,
				Reps:               s.Reps,
				RPE:                s.RPE,
				DistanceM:          s.DistanceMeters,
				DurationS:          s.DurationSeconds,
			}
			if set.Working() && set.WeightKg != nil {
				w.VolumeKg += *set.WeightKg * float64(*set.Reps)
			}
			w.Sets = append(w.Sets, set)
		}
	}
	w.VolumeKg = math.Round(w.VolumeKg*10) / 10
	return w, nil
}
