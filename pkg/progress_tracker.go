package pkg

import (
	"sync/atomic"
	"time"
)

type ProgressTracker struct {
	phase          string
	startTime      time.Time
	totalItems     int
	completedItems int64
	callback       ProgressCallback
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(phase string, totalItems int, callback ProgressCallback) *ProgressTracker {
	tracker := &ProgressTracker{
		phase:      phase,
		startTime:  time.Now(),
		totalItems: totalItems,
		callback:   callback,
	}

	// Report initial progress
	if callback != nil {
		callback(ProgressInfo{
			Phase:      phase,
			Current:    "Initializing...",
			Completed:  0,
			Total:      totalItems,
			Percentage: 0.0,
			StartTime:  tracker.startTime,
		})
	}

	return tracker
}

// UpdateProgress atomically increments progress and reports it
func (pt *ProgressTracker) UpdateProgress(currentItem string) {
	if pt == nil || pt.callback == nil {
		return
	}

	completed := atomic.AddInt64(&pt.completedItems, 1)
	percentage := float64(completed) / float64(pt.totalItems) * 100.0

	pt.callback(ProgressInfo{
		Phase:      pt.phase,
		Current:    currentItem,
		Completed:  int(completed),
		Total:      pt.totalItems,
		Percentage: percentage,
		StartTime:  pt.startTime,
	})
}

// Complete reports completion
func (pt *ProgressTracker) Complete() {
	if pt == nil || pt.callback == nil {
		return
	}

	pt.callback(ProgressInfo{
		Phase:      pt.phase,
		Current:    "Completed",
		Completed:  pt.totalItems,
		Total:      pt.totalItems,
		Percentage: 100.0,
		StartTime:  pt.startTime,
	})
}
