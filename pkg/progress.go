package pkg

import (
	"fmt"
	"strings"
	"time"
)

// ProgressInfo represents progress information for a long-running operation
type ProgressInfo struct {
	Phase      string    // "scanning" or "indexing"
	Current    string    // Current item being processed
	Completed  int       // Number of items completed
	Total      int       // Total number of items
	Percentage float64   // Completion percentage (0-100)
	StartTime  time.Time // When the operation started
}

// ProgressCallback is called to report progress updates
type ProgressCallback func(ProgressInfo)

// createProgressBar generates a Unicode progress bar string
func createProgressBar(percentage float64, width int) string {
	if width <= 0 {
		width = 50
	}
	
	filled := int(float64(width) * percentage / 100.0)
	bar := ""
	
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	
	return bar
}

// calculateETA estimates time remaining based on current progress
func calculateETA(elapsed time.Duration, percentage float64) time.Duration {
	if percentage <= 0 {
		return 0
	}
	
	totalEstimated := elapsed * time.Duration(100.0/percentage)
	eta := totalEstimated - elapsed
	
	if eta < 0 {
		return 0
	}
	
	return eta
}

// calculateProcessingRate calculates items processed per second
func calculateProcessingRate(completed int, elapsed time.Duration) float64 {
	if elapsed.Seconds() <= 0 {
		return 0
	}
	
	return float64(completed) / elapsed.Seconds()
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	if maxLen <= 3 {
		return s[:maxLen]
	}
	
	return s[:maxLen-3] + "..."
}

// CreateRichProgressCallback creates a callback that displays rich progress information
func CreateRichProgressCallback() ProgressCallback {
	return func(progress ProgressInfo) {
		elapsed := time.Since(progress.StartTime)
		
		// Calculate ETA
		var eta time.Duration
		if progress.Percentage > 0 && progress.Percentage < 100 {
			eta = calculateETA(elapsed, progress.Percentage)
		}
		
		// Calculate processing rate
		rate := calculateProcessingRate(progress.Completed, elapsed)
		
		// Create progress bar
		bar := createProgressBar(progress.Percentage, 50)
		
		// Truncate current item name if too long
		current := truncateString(progress.Current, 30)
		
		// Display rich progress with Unicode indicators
		fmt.Printf("\r🔄 %s | [%s] %.1f%% (%d/%d) | ⏱️ %.1fs",
			strings.Title(progress.Phase), bar, progress.Percentage, 
			progress.Completed, progress.Total, elapsed.Seconds())
		
		if progress.Percentage > 0 && progress.Percentage < 100 {
			fmt.Printf(" | 🔮 ETA: %.1fs", eta.Seconds())
		}
		
		if current != "" && current != "Completed" {
			fmt.Printf(" | 📦 %s", current)
		}
		
		if rate > 0 {
			fmt.Printf(" | ⚡ %.1f/s", rate)
		}
		
		if progress.Percentage >= 100 {
			fmt.Printf("\n✅ %s completed!\n", strings.Title(progress.Phase))
		}
	}
}

// CreateSimpleProgressCallback creates a simple progress callback for basic output
func CreateSimpleProgressCallback() ProgressCallback {
	return func(progress ProgressInfo) {
		if progress.Percentage >= 100 {
			fmt.Printf("✅ %s completed: %d/%d items\n", 
				strings.Title(progress.Phase), progress.Completed, progress.Total)
		} else {
			fmt.Printf("🔄 %s: %.1f%% (%d/%d) - %s\n", 
				strings.Title(progress.Phase), progress.Percentage, 
				progress.Completed, progress.Total, progress.Current)
		}
	}
}
