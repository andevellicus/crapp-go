package repository

import (
	"database/sql"
	"math/rand"
	"time"

	"crapp-go/internal/database"
	"crapp-go/internal/models"

	"github.com/lib/pq"
)

func GetOrCreateAssessmentState(userID int, totalQuestions int) (*models.AssessmentState, error) {
	// First, try to find an existing active assessment.
	state := &models.AssessmentState{}
	selectQuery := `
		SELECT id, user_id, is_complete, question_order, current_question_index
		FROM assessments
		WHERE user_id = $1 AND is_complete = false
		LIMIT 1
	`
	err := database.DB.QueryRow(selectQuery, userID).Scan(&state.ID, &state.UserID, &state.IsComplete, &state.QuestionOrder, &state.CurrentQuestionIndex)

	// If we found an existing one, return it.
	if err == nil {
		return state, nil
	}

	// If no row was found, and we have questions to create an assessment with, create one.
	if err == sql.ErrNoRows && totalQuestions > 0 {
		order := make([]int, totalQuestions)
		for i := range order {
			order[i] = i
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

		order64 := make([]int64, len(order))
		for i, v := range order {
			order64[i] = int64(v)
		}

		insertQuery := `
			INSERT INTO assessments (user_id, question_order, current_question_index)
			VALUES ($1, $2, 0)
			RETURNING id, user_id, is_complete, question_order, current_question_index
		`

		err = database.DB.QueryRow(insertQuery, userID, pq.Array(order64)).Scan(&state.ID, &state.UserID, &state.IsComplete, &state.QuestionOrder, &state.CurrentQuestionIndex)
		return state, err
	}

	// If there was an error other than "no rows", or if we were asked to create an assessment with no questions, return the error.
	return nil, err
}

func UpdateAssessmentIndex(assessmentID int, newIndex int) error {
	query := `UPDATE assessments SET current_question_index = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := database.DB.Exec(query, newIndex, assessmentID)
	return err
}

func CompleteAssessment(assessmentID int) error {
	query := `UPDATE assessments SET is_complete = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := database.DB.Exec(query, assessmentID)
	return err
}

func GetAnswersForAssessment(assessmentID int) (map[string]string, error) {
	rows, err := database.DB.Query(`SELECT question_id, answer_value FROM answers WHERE assessment_id = $1`, assessmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answers := make(map[string]string)
	for rows.Next() {
		var questionID, answerValue string
		if err := rows.Scan(&questionID, &answerValue); err != nil {
			return nil, err
		}
		answers[questionID] = answerValue
	}
	return answers, nil
}
