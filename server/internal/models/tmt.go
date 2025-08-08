package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// TMTResult holds the processed metrics from a Trail Making Test.
type TMTResult struct {
	gorm.Model
	AssessmentID        uint
	Assessment          AssessmentState `gorm:"foreignKey:AssessmentID"`
	PartACompletionTime float64
	PartAErrors         int
	PartBCompletionTime float64
	PartBErrors         int
	BToARatio           float64
	RawData             json.RawMessage `gorm:"type:jsonb"`
	CreatedAt           time.Time
}

// TMTClick represents a single click event during a Trail Making Test.
type TMTClick struct {
	gorm.Model
	ResultID    uint
	Result      TMTResult `gorm:"foreignKey:ResultID"`
	X           float64
	Y           float64
	Time        float64
	TargetItem  int
	CurrentPart string
}
