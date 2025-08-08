// server/internal/repository/results.go
package repository

import (
	"crapp-go/internal/database"
	"crapp-go/internal/metrics" // Import metrics package
	"crapp-go/internal/models"

	"gorm.io/gorm"
)

// SaveAnswer saves a standard answer (text, radio, dropdown) to the database.
func SaveAnswer(assessmentID uint, questionID string, answerValue string) error {
	answer := models.Answer{
		AssessmentID: assessmentID,
		QuestionID:   questionID,
		AnswerValue:  answerValue,
	}
	return database.DB.Where(models.Answer{AssessmentID: assessmentID, QuestionID: questionID}).Assign(answer).FirstOrCreate(&answer).Error
}

// SaveCPTResultTx saves the summary and all granular events for a CPT test in a single transaction.
func SaveCPTResultTx(summary models.CPTResult, events []models.CPTEvent) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&summary).Error; err != nil {
			return err
		}
		for i := range events {
			events[i].ResultID = summary.ID
		}
		if err := tx.Create(&events).Error; err != nil {
			return err
		}
		return nil
	})
}

// SaveDSTResultTx saves the summary and all attempts for a DST test in a single transaction.
func SaveDSTResultTx(summary models.DSTResult, attempts []metrics.DigitSpanAttempt) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&summary).Error; err != nil {
			return err
		}
		dstAttempts := make([]models.DSTAttempt, len(attempts))
		for i, attempt := range attempts {
			dstAttempts[i] = models.DSTAttempt{
				ResultID:  summary.ID,
				Span:      attempt.Span,
				Trial:     attempt.Trial,
				Sequence:  attempt.Sequence,
				Input:     attempt.Input,
				IsCorrect: attempt.Correct,
				Timestamp: attempt.Timestamp,
			}
		}
		if err := tx.Create(&dstAttempts).Error; err != nil {
			return err
		}
		return nil
	})
}

// SaveTMTResultTx saves the summary and all clicks for a TMT test in a single transaction.
func SaveTMTResultTx(summary models.TMTResult, clicks []metrics.Click) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&summary).Error; err != nil {
			return err
		}
		tmtClicks := make([]models.TMTClick, len(clicks))
		for i, click := range clicks {
			tmtClicks[i] = models.TMTClick{
				ResultID:    summary.ID,
				X:           click.X,
				Y:           click.Y,
				Time:        click.Time,
				TargetItem:  click.TargetItem,
				CurrentPart: click.CurrentPart,
			}
		}
		if err := tx.Create(&tmtClicks).Error; err != nil {
			return err
		}
		return nil
	})
}
