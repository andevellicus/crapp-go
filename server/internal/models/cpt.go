package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

// CPTResult holds the processed metrics from a CPT test.
type CPTResult struct {
	gorm.Model
	AssessmentID        uint
	Assessment          AssessmentState `gorm:"foreignKey:AssessmentID"`
	CorrectDetections   int
	CommissionErrors    int
	OmissionErrors      int
	AverageReactionTime float64
	ReactionTimeSD      float64
	DetectionRate       float64
	OmissionErrorRate   float64
	CommissionErrorRate float64
	RawData             json.RawMessage `gorm:"type:jsonb"`
}

// CPTEvent represents a single event (stimulus or response) in a CPT test.
type CPTEvent struct {
	gorm.Model
	ResultID      uint
	Result        CPTResult `gorm:"foreignKey:ResultID"`
	EventType     string    // 'stimulus' or 'response'
	StimulusValue *string   // Pointer to allow null
	IsTarget      *bool     // Pointer to allow null
	PresentedAt   *float64
	ResponseTime  *float64
	StimulusIndex *int
}
