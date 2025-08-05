package repository

import (
	"math/rand"
	"time"

	"crapp-go/internal/database"
	"crapp-go/internal/models"

	"github.com/lib/pq"
)

func GetOrCreateAssessmentState(userID int, totalQuestions int) (*models.AssessmentState, error) {
	state := &models.AssessmentState{}
	selectQuery := `
		SELECT id, user_id, is_complete, question_order, current_question_index
		FROM assessments
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`
	err := database.DB.QueryRow(selectQuery, userID).Scan(&state.ID, &state.UserID, &state.IsComplete, &state.QuestionOrder, &state.CurrentQuestionIndex)

	// Case 1: An assessment was found.
	if err == nil {
		// If it's not complete, or if we are just fetching state without the intent to create, return it.
		if !state.IsComplete || totalQuestions <= 0 {
			return state, nil
		}
		// If it IS complete and we want to start a new one, proceed to the creation logic below.
	}

	// Case 2: No assessment was found (err is sql.ErrNoRows) or the last one was complete.
	// We should create a new one if requested.
	if totalQuestions > 0 {
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

		newState := &models.AssessmentState{}
		insertErr := database.DB.QueryRow(insertQuery, userID, pq.Array(order64)).Scan(&newState.ID, &newState.UserID, &newState.IsComplete, &newState.QuestionOrder, &newState.CurrentQuestionIndex)
		return newState, insertErr
	}

	// Case 3: No assessment found and not asked to create one, or another DB error occurred.
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
