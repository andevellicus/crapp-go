package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// DSTResult holds the processed metrics from a Digit Span test.
type DSTResult struct {
	gorm.Model
	AssessmentID        uint
	Assessment          AssessmentState `gorm:"foreignKey:AssessmentID"`
	HighestSpanAchieved int
	TotalTrials         int
	CorrectTrials       int
	RawData             json.RawMessage `gorm:"type:jsonb"`
	CreatedAt           time.Time
}

// DSTAttempt represents a single trial within a Digit Span Test.
type DSTAttempt struct {
	gorm.Model
	ResultID  uint
	Result    DSTResult `gorm:"foreignKey:ResultID"`
	Span      int
	Trial     int
	Sequence  string
	Input     string
	IsCorrect bool
	Timestamp float64
}
