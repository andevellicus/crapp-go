package models

import (
	"encoding/json"
	"time"
)

// CPTResult holds the processed metrics from a CPT test.
type CPTResult struct {
	ID                  int
	AssessmentID        int
	CorrectDetections   int
	CommissionErrors    int
	OmissionErrors      int
	AverageReactionTime float64
	ReactionTimeSD      float64
	DetectionRate       float64
	OmissionErrorRate   float64
	CommissionErrorRate float64
	RawData             json.RawMessage `gorm:"type:jsonb"`
	CreatedAt           time.Time
}

// CPTEvent represents a single event (stimulus or response) in a CPT test.
type CPTEvent struct {
	ID            int
	ResultID      int
	EventType     string  // 'stimulus' or 'response'
	StimulusValue *string // Pointer to allow null
	IsTarget      *bool   // Pointer to allow null
	PresentedAt   *float64
	ResponseTime  *float64
	StimulusIndex *int
}
