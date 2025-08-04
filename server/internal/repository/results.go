// server/internal/repository/results.go
package repository

import (
	"context"
	"crapp-go/internal/database"
	"crapp-go/internal/metrics" // Import metrics package
	"crapp-go/internal/models"
)

// SaveAnswer saves a standard answer (text, radio, dropdown) to the database.
func SaveAnswer(assessmentID int, questionID string, answerValue string) error {
	query := `INSERT INTO answers (assessment_id, question_id, answer_value) VALUES ($1, $2, $3) ON CONFLICT (assessment_id, question_id) DO UPDATE SET answer_value = EXCLUDED.answer_value;`
	_, err := database.DB.Exec(query, assessmentID, questionID, answerValue)
	return err
}

// SaveCPTResultTx saves the summary and all granular events for a CPT test in a single transaction.
func SaveCPTResultTx(summary models.CPTResult, events []models.CPTEvent) error {
	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert summary and get its ID
	summaryQuery := `INSERT INTO cpt_results (assessment_id, correct_detections, commission_errors, omission_errors, average_reaction_time, reaction_time_sd, detection_rate, omission_error_rate, commission_error_rate, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	var resultID int
	err = tx.QueryRow(summaryQuery, summary.AssessmentID, summary.CorrectDetections, summary.CommissionErrors, summary.OmissionErrors, summary.AverageReactionTime, summary.ReactionTimeSD, summary.DetectionRate, summary.OmissionErrorRate, summary.CommissionErrorRate, summary.CreatedAt).Scan(&resultID)
	if err != nil {
		return err
	}

	// 2. Insert all granular events referencing the summary ID
	stmt, err := tx.Prepare(`INSERT INTO cpt_events (result_id, event_type, stimulus_value, is_target, presented_at, response_time, stimulus_index) VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		_, err = stmt.Exec(resultID, event.EventType, event.StimulusValue, event.IsTarget, event.PresentedAt, event.ResponseTime, event.StimulusIndex)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SaveDSTResultTx saves the summary and all attempts for a DST test in a single transaction.
func SaveDSTResultTx(summary models.DSTResult, attempts []metrics.DigitSpanAttempt) error {
	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	summaryQuery := `INSERT INTO dst_results (assessment_id, highest_span_achieved, total_trials, correct_trials, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var resultID int
	err = tx.QueryRow(summaryQuery, summary.AssessmentID, summary.HighestSpanAchieved, summary.TotalTrials, summary.CorrectTrials, summary.CreatedAt).Scan(&resultID)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO dst_attempts (result_id, span, trial, sequence, input, is_correct, "timestamp") VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, attempt := range attempts {
		_, err = stmt.Exec(resultID, attempt.Span, attempt.Trial, attempt.Sequence, attempt.Input, attempt.Correct, attempt.Timestamp)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SaveTMTResultTx saves the summary and all clicks for a TMT test in a single transaction.
func SaveTMTResultTx(summary models.TMTResult, clicks []metrics.Click) error {
	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	summaryQuery := `INSERT INTO tmt_results (assessment_id, part_a_completion_time, part_a_errors, part_b_completion_time, part_b_errors, b_to_a_ratio, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var resultID int
	err = tx.QueryRow(summaryQuery, summary.AssessmentID, summary.PartACompletionTime, summary.PartAErrors, summary.PartBCompletionTime, summary.PartBErrors, summary.BToARatio, summary.CreatedAt).Scan(&resultID)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO tmt_clicks (result_id, x, y, "time", target_item, current_part) VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, click := range clicks {
		_, err = stmt.Exec(resultID, click.X, click.Y, click.Time, click.TargetItem, click.CurrentPart)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
