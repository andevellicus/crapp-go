package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type AssessmentState struct {
	ID                   int `gorm:"primaryKey"`
	UserID               int
	User                 User `gorm:"foreignKey:UserID"`
	IsComplete           bool
	QuestionOrder        pq.Int64Array `gorm:"type:integer[]"`
	CurrentQuestionIndex int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Need to define a new model for the answers table
type Answer struct {
	gorm.Model
	AssessmentID uint
	Assessment   AssessmentState `gorm:"foreignKey:AssessmentID"`
	QuestionID   string
	AnswerValue  string
}
