package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateProgressBar(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		width      int
		expected   string
	}{
		{
			name:       "empty progress bar",
			percentage: 0,
			width:      10,
			expected:   "░░░░░░░░░░",
		},
		{
			name:       "half full progress bar",
			percentage: 50,
			width:      10,
			expected:   "█████░░░░░",
		},
		{
			name:       "full progress bar",
			percentage: 100,
			width:      10,
			expected:   "██████████",
		},
		{
			name:       "default width when zero",
			percentage: 25,
			width:      0,
			expected:   "████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░",
		},
		{
			name:       "partial progress",
			percentage: 33.3,
			width:      6,
			expected:   "█░░░░░",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createProgressBar(tt.percentage, tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateETA(t *testing.T) {
	tests := []struct {
		name       string
		elapsed    time.Duration
		percentage float64
		expected   time.Duration
	}{
		{
			name:       "50% complete",
			elapsed:    10 * time.Second,
			percentage: 50.0,
			expected:   10 * time.Second, // Total: 20s, Remaining: 10s
		},
		{
			name:       "25% complete",
			elapsed:    5 * time.Second,
			percentage: 25.0,
			expected:   15 * time.Second, // Total: 20s, Remaining: 15s
		},
		{
			name:       "0% complete",
			elapsed:    5 * time.Second,
			percentage: 0.0,
			expected:   0 * time.Second, // No progress, can't estimate
		},
		{
			name:       "100% complete",
			elapsed:    10 * time.Second,
			percentage: 100.0,
			expected:   0 * time.Second, // Already done
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateETA(tt.elapsed, tt.percentage)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateProcessingRate(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		elapsed   time.Duration
		expected  float64
	}{
		{
			name:      "10 items in 10 seconds",
			completed: 10,
			elapsed:   10 * time.Second,
			expected:  1.0,
		},
		{
			name:      "50 items in 10 seconds",
			completed: 50,
			elapsed:   10 * time.Second,
			expected:  5.0,
		},
		{
			name:      "0 items processed",
			completed: 0,
			elapsed:   10 * time.Second,
			expected:  0.0,
		},
		{
			name:      "no time elapsed",
			completed: 10,
			elapsed:   0 * time.Second,
			expected:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateProcessingRate(tt.completed, tt.elapsed)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string with ellipsis",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is...",
		},
		{
			name:     "very short maxLen",
			input:    "hello",
			maxLen:   3,
			expected: "hel",
		},
		{
			name:     "maxLen 1",
			input:    "hello",
			maxLen:   1,
			expected: "h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProgressCallback(t *testing.T) {
	// Test that callbacks don't panic and can handle various progress states
	richCallback := CreateRichProgressCallback()
	simpleCallback := CreateSimpleProgressCallback()

	testCases := []ProgressInfo{
		{
			Phase:      "testing",
			Current:    "test_item_1",
			Completed:  0,
			Total:      100,
			Percentage: 0.0,
			StartTime:  time.Now().Add(-1 * time.Second),
		},
		{
			Phase:      "testing",
			Current:    "test_item_50",
			Completed:  50,
			Total:      100,
			Percentage: 50.0,
			StartTime:  time.Now().Add(-10 * time.Second),
		},
		{
			Phase:      "testing",
			Current:    "Completed",
			Completed:  100,
			Total:      100,
			Percentage: 100.0,
			StartTime:  time.Now().Add(-20 * time.Second),
		},
	}

	// Test that callbacks don't panic
	for _, tc := range testCases {
		assert.NotPanics(t, func() {
			richCallback(tc)
		})
		
		assert.NotPanics(t, func() {
			simpleCallback(tc)
		})
	}
}
