package metrics

import (
	"encoding/json"
	"time"

	"crapp-go/internal/models"
)

// Trail Making Test results processing

// TrailMakingData represents the raw data from a Trail Making Test
type TrailMakingData struct {
	TestStartTime       float64        `json:"testStartTime"`
	TestEndTime         float64        `json:"testEndTime"`
	PartAStartTime      float64        `json:"partAStartTime"`
	PartAEndTime        float64        `json:"partAEndTime"`
	PartBStartTime      float64        `json:"partBStartTime"`
	PartBEndTime        float64        `json:"partBEndTime"`
	PartAErrors         int            `json:"partAErrors"`
	PartBErrors         int            `json:"partBErrors"`
	PartACompletionTime float64        `json:"partACompletionTime"`
	PartBCompletionTime float64        `json:"partBCompletionTime"`
	Clicks              []Click        `json:"clicks"`
	Settings            map[string]any `json:"settings"`
}

// Click represents a single interaction during the Trail Making Test
type Click struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Time        float64 `json:"time"`
	TargetItem  int     `json:"targetItem"`
	CurrentPart string  `json:"currentPart"`
}

// Calculate metrics for Trail Making Test
func CalculateTrailMetrics(data *TrailMakingData) *models.TMTResult {
	// Create Trail Making Test result model
	return &models.TMTResult{
		// Time fields
		//TestStartTime: time.UnixMilli(int64(data.TestStartTime)),
		//TestEndTime:   time.UnixMilli(int64(data.TestEndTime)),

		// Part A metrics
		PartACompletionTime: data.PartACompletionTime,
		PartAErrors:         data.PartAErrors,

		// Part B metrics
		PartBCompletionTime: data.PartBCompletionTime,
		PartBErrors:         data.PartBErrors,

		// Calculated metrics
		BToARatio: calculateBToARatio(data),

		// Store the raw data for future analysis
		RawData:   serializeTrailData(data),
		CreatedAt: time.Now(),
	}
}

// Calculate B/A ratio (important clinical measure)
func calculateBToARatio(data *TrailMakingData) float64 {
	if data.PartACompletionTime <= 0 {
		return 0
	}
	return data.PartBCompletionTime / data.PartACompletionTime
}

// Serialize trail data to JSON
func serializeTrailData(data *TrailMakingData) json.RawMessage {
	result, err := json.Marshal(data)
	if err != nil {
		return json.RawMessage("{}")
	}
	return result
}
