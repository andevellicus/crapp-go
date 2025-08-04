package metrics

import (
	"math"
)

type CPTStimulusPresentation struct {
	Value       string  `json:"value"`
	IsTarget    bool    `json:"isTarget"`
	PresentedAt float64 `json:"presentedAt"`
}

type CPTResponse struct {
	Stimulus      string  `json:"stimulus"`
	IsTarget      bool    `json:"isTarget"`
	ResponseTime  float64 `json:"responseTime"`
	StimulusIndex int     `json:"stimulusIndex"`
}

// CPTData represents the structure of raw CPT test data
type CPTData struct {
	TestStartTime    float64                   `json:"testStartTime"`
	TestEndTime      float64                   `json:"testEndTime"`
	StimuliPresented []CPTStimulusPresentation `json:"stimuliPresented"`
	Responses        []CPTResponse             `json:"responses"`
	Settings         map[string]any            `json:"settings"`
}

// Helper methods for CPT calculations
func CountCorrectDetections(data *CPTData) int {
	count := 0
	for _, response := range data.Responses {
		if response.IsTarget {
			count++
		}
	}
	return count
}

func CountCommissionErrors(data *CPTData) int {
	count := 0
	for _, response := range data.Responses {
		if !response.IsTarget {
			count++
		}
	}
	return count
}

func CountOmissionErrors(data *CPTData) int {
	// Count total targets presented
	totalTargets := 0
	for _, stim := range data.StimuliPresented {
		if stim.IsTarget {
			totalTargets++
		}
	}

	// Omission errors = targets missed
	return totalTargets - CountCorrectDetections(data)
}

func CalculateAverageReactionTime(data *CPTData) float64 {
	var sum float64
	var count int

	for _, response := range data.Responses {
		if response.IsTarget {
			sum += response.ResponseTime
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func CalculateReactionTimeSD(data *CPTData) float64 {
	// Get reaction times and average
	var reactionTimes []float64
	for _, response := range data.Responses {
		if response.IsTarget {
			reactionTimes = append(reactionTimes, response.ResponseTime)
		}
	}

	if len(reactionTimes) <= 1 {
		return 0
	}

	avg := CalculateAverageReactionTime(data)
	var sumSquaredDiff float64

	for _, rt := range reactionTimes {
		diff := rt - avg
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(reactionTimes))
	return math.Sqrt(variance)
}

func CalculateDetectionRate(data *CPTData) float64 {
	// Count total targets presented
	totalTargets := 0
	for _, stim := range data.StimuliPresented {
		if stim.IsTarget {
			totalTargets++
		}
	}

	if totalTargets == 0 {
		return 0
	}

	return float64(CountCorrectDetections(data)) / float64(totalTargets)
}

func CalculateOmissionErrorRate(data *CPTData) float64 {
	// Count total targets presented
	totalTargets := 0
	for _, stim := range data.StimuliPresented {
		if stim.IsTarget {
			totalTargets++
		}
	}

	if totalTargets == 0 {
		return 0
	}

	return float64(CountOmissionErrors(data)) / float64(totalTargets)
}

func CalculateCommissionErrorRate(data *CPTData) float64 {
	// Count non-targets
	nonTargetCount := 0
	for _, stim := range data.StimuliPresented {
		if !stim.IsTarget {
			nonTargetCount++
		}
	}

	if nonTargetCount == 0 {
		return 0
	}

	return float64(CountCommissionErrors(data)) / float64(nonTargetCount)
}
