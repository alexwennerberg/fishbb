package main

import (
	"testing"
	"time"
)

func TestTimeago(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "1 second ago",
			time:     now.Add(-1 * time.Second),
			expected: "1 second ago",
		},
		{
			name:     "30 seconds ago",
			time:     now.Add(-30 * time.Second),
			expected: "30 seconds ago",
		},
		{
			name:     "1 minute ago",
			time:     now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "45 minutes ago",
			time:     now.Add(-45 * time.Minute),
			expected: "45 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "12 hours ago",
			time:     now.Add(-12 * time.Hour),
			expected: "12 hours ago",
		},
		{
			name:     "1 day ago",
			time:     now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "7 days ago",
			time:     now.Add(-7 * 24 * time.Hour),
			expected: "7 days ago",
		},
		{
			name:     "30 days ago",
			time:     now.Add(-30 * 24 * time.Hour),
			expected: "30 days ago",
		},
		{
			name:     "365 days ago",
			time:     now.Add(-365 * 24 * time.Hour),
			expected: "365 days ago",
		},
		{
			name:     "more than 1 year ago",
			time:     time.Date(2020, 6, 15, 12, 0, 0, 0, time.UTC),
			expected: "2020-06-15",
		},
		{
			name:     "2 years ago",
			time:     now.Add(-2 * 365 * 24 * time.Hour),
			expected: now.Add(-2 * 365 * 24 * time.Hour).Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeago(&tt.time)
			if result != tt.expected {
				t.Errorf("timeago() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTimeagoSingularPlural(t *testing.T) {
	now := time.Now()

	// Test singular forms
	singularTests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"1 second", 1 * time.Second, "1 second ago"},
		{"1 minute", 1 * time.Minute, "1 minute ago"},
		{"1 hour", 1 * time.Hour, "1 hour ago"},
		{"1 day", 24 * time.Hour, "1 day ago"},
	}

	for _, tt := range singularTests {
		t.Run(tt.name, func(t *testing.T) {
			pastTime := now.Add(-tt.duration)
			result := timeago(&pastTime)
			if result != tt.contains {
				t.Errorf("timeago() = %v, expected %v", result, tt.contains)
			}
		})
	}

	// Test plural forms
	pluralTests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"2 seconds", 2 * time.Second, "2 seconds ago"},
		{"2 minutes", 2 * time.Minute, "2 minutes ago"},
		{"2 hours", 2 * time.Hour, "2 hours ago"},
		{"2 days", 48 * time.Hour, "2 days ago"},
	}

	for _, tt := range pluralTests {
		t.Run(tt.name, func(t *testing.T) {
			pastTime := now.Add(-tt.duration)
			result := timeago(&pastTime)
			if result != tt.contains {
				t.Errorf("timeago() = %v, expected %v", result, tt.contains)
			}
		})
	}
}
