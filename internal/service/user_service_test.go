package service

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day int) time.Time {
	return time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
}

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name string
		dob  time.Time
		now  time.Time
		want int
	}{
		{"birthday passed", d(1990, 5, 10), d(2024, 6, 15), 34},
		{"birthday today", d(1990, 6, 15), d(2024, 6, 15), 34},
		{"birthday tomorrow", d(1990, 6, 16), d(2024, 6, 15), 33},
		{"leap day, day before", d(2000, 2, 29), d(2023, 2, 28), 22},
		{"leap day, after", d(2000, 2, 29), d(2023, 3, 1), 23},
		{"born today", d(2024, 6, 15), d(2024, 6, 15), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CalculateAge(tc.dob, tc.now); got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}