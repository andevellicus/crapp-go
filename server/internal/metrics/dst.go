package metrics

import (
	"crapp-go/internal/models"
)

type DigitSpanAttempt struct {
	Span      int     `json:"span"`
	Trial     int     `json:"trial"`
	Sequence  string  `json:"sequence"`
	Input     string  `json:"input"`
	Correct   bool    `json:"correct"`
	Timestamp float64 `json:"timestamp"` // Relative timestamp from test start
}

type DigitSpanRawData struct {
	TestStartTime float64            `json:"testStartTime"` // JS performance.now() timestamp
	TestEndTime   float64            `json:"testEndTime"`   // JS performance.now() timestamp
	Results       []DigitSpanAttempt `json:"results"`       // Array of attempt data
	Settings      map[string]any     `json:"settings"`      // Test settings used
}

func CalculateDigitSpanMetrics(results *DigitSpanRawData) (*models.DSTResult, error) {
	// --- Calculate Metrics ---
	highestSpan := 0
	totalTrials := len(results.Results)
	correctTrials := 0
	initialSpan := 3 // Default

	// Safely get initialSpan from settings
	if settingsInitialSpan, ok := results.Settings["initialSpan"]; ok {
		if val, ok := settingsInitialSpan.(float64); ok { // JSON numbers often float64
			initialSpan = int(val)
		}
	}
	highestSpan = initialSpan - 1 // Start assuming failure at initial span

	hasCorrectAttempts := false
	minAttemptedSpan := initialSpan // Track the lowest span actually attempted

	for _, attempt := range results.Results {
		if attempt.Span < minAttemptedSpan {
			minAttemptedSpan = attempt.Span
		}
		if attempt.Correct {
			correctTrials++
			hasCorrectAttempts = true
			if attempt.Span > highestSpan {
				highestSpan = attempt.Span
			}
		}
	}

	// If no correct attempts at all, highest span is one less than the minimum attempted span
	if !hasCorrectAttempts && totalTrials > 0 {
		highestSpan = minAttemptedSpan - 1
	}
	// Ensure span doesn't go below 0
	if highestSpan < 0 {
		highestSpan = 0
	}

	// --- Create the Result Object (partially populated) ---
	result := &models.DSTResult{
		HighestSpanAchieved: highestSpan,
		TotalTrials:         totalTrials,
		CorrectTrials:       correctTrials,
	}
	return result, nil
}
