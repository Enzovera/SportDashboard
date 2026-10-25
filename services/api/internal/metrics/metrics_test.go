package metrics

import (
	"math"
	"testing"
	"time"
)

func d(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func f(v float64) *float64 { return &v }

func TestWeekStart(t *testing.T) {
	tests := []struct{ in, want string }{
		{"2026-10-25", "2026-10-19"}, // Sunday
		{"2026-10-19", "2026-10-19"}, // Monday
		{"2026-10-21", "2026-10-19"},
		{"2026-11-01", "2026-10-26"},
	}
	for _, tt := range tests {
		if got := WeekStart(d(tt.in)); !got.Equal(d(tt.want)) {
			t.Errorf("WeekStart(%s) = %s, want %s", tt.in, got.Format(time.DateOnly), tt.want)
		}
	}
}

func TestLoadSeries(t *testing.T) {
	// Constant daily load: both averages approach it, ATL (7 d) much faster than CTL (42 d).
	daily := map[time.Time]float64{}
	from, to := d("2026-01-01"), d("2026-06-30")
	for _, x := range Days(from, to) {
		daily[x] = 100
	}
	s := LoadSeries(daily, from, to)
	last := s[len(s)-1]
	// After 180 days CTL is within e^(-180/42) ≈ 1.4% of the steady state.
	if math.Abs(last.ATL-100) > 0.01 || math.Abs(last.CTL-100) > 2 {
		t.Errorf("steady state ATL=%.2f CTL=%.2f", last.ATL, last.CTL)
	}
	if s[6].ATL <= s[6].CTL {
		t.Errorf("after a week ATL (%.1f) should lead CTL (%.1f)", s[6].ATL, s[6].CTL)
	}

	// One hard day then rest: TSB the morning after is negative; fatigue (7 d) decays faster than fitness
	// (42 d), so TSB turns positive after ~15 days (300·α7·e^(-t/7) = 300·α42·e^(-t/42)).
	spike := map[time.Time]float64{d("2026-10-01"): 300}
	s = LoadSeries(spike, d("2026-09-01"), d("2026-10-30"))
	if tsb, _ := TSB(s, d("2026-10-02")); tsb >= 0 {
		t.Errorf("TSB after spike = %.1f, want < 0", tsb)
	}
	if tsb, _ := TSB(s, d("2026-10-25")); tsb <= 0 {
		t.Errorf("TSB three weeks later = %.1f, want > 0", tsb)
	}
	if _, ok := TSB(s, d("2026-09-01")); ok {
		t.Error("TSB on the first day has no previous day")
	}
}

func TestFormOf(t *testing.T) {
	tests := []struct {
		tsb  float64
		want Form
	}{{4, Fresh}, {0, Fresh}, {-0.5, Neutral}, {-10, Neutral}, {-10.1, Tired}, {-30, Tired}, {-31, Overreached}}
	for _, tt := range tests {
		if got := FormOf(tt.tsb); got != tt.want {
			t.Errorf("FormOf(%v) = %s, want %s", tt.tsb, got, tt.want)
		}
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		in   []*float64
		min  int
		want float64
		ok   bool
	}{
		{[]*float64{f(50), nil, f(54)}, 1, 52, true},
		{[]*float64{f(50), nil, nil}, 2, 0, false},
		{nil, 1, 0, false},
	}
	for _, tt := range tests {
		got, ok := Mean(tt.in, tt.min)
		if ok != tt.ok || got != tt.want {
			t.Errorf("Mean = %v,%v want %v,%v", got, ok, tt.want, tt.ok)
		}
	}
}
