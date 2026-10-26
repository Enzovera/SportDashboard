package syncreq

import (
	"errors"
	"testing"
	"time"
)

func TestCheckCooldown(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	ago := func(d time.Duration) *time.Time { t := now.Add(-d); return &t }
	tests := []struct {
		name  string
		last  *time.Time
		retry time.Duration // 0 = allowed
	}{
		{"never synced", nil, 0},
		{"long ago", ago(time.Hour), 0},
		{"exactly cooldown", ago(Cooldown), 0},
		{"just finished", ago(0), Cooldown},
		{"30 s ago", ago(30 * time.Second), 90 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkCooldown(tt.last, now)
			var tooSoon TooSoonError
			switch {
			case tt.retry == 0 && err != nil:
				t.Fatalf("want allowed, got %v", err)
			case tt.retry != 0 && !errors.As(err, &tooSoon):
				t.Fatalf("want TooSoonError, got %v", err)
			case tt.retry != 0 && tooSoon.RetryAfter != tt.retry:
				t.Fatalf("RetryAfter = %s, want %s", tooSoon.RetryAfter, tt.retry)
			}
		})
	}
}
