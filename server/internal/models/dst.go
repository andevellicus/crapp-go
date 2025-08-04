package models

import (
	"encoding/json"
	"time"
)

// DSTResult holds the processed metrics from a Digit Span test.
type DSTResult struct {
	ID                  int
	AssessmentID        int
	HighestSpanAchieved int
	TotalTrials         int
	CorrectTrials       int
	RawData             json.RawMessage `gorm:"type:jsonb"`
	CreatedAt           time.Time
}

// DSTAttempt represents a single trial within a Digit Span Test.
type DSTAttempt struct {
	ID        int
	ResultID  int
	Span      int
	Trial     int
	Sequence  string
	Input     string
	IsCorrect bool
	Timestamp float64
}
