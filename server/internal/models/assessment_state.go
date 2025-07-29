package models

import (
	"time"

	"github.com/lib/pq"
)

type AssessmentState struct {
	ID                   int
	UserID               int
	IsComplete           bool
	QuestionOrder        pq.Int64Array `gorm:"type:integer[]"`
	CurrentQuestionIndex int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
