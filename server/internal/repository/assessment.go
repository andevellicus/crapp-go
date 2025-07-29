package repository

import (
	"math/rand"
	"time"

	"crapp-go/internal/database"
	"crapp-go/internal/models"

	"github.com/lib/pq"
)

func GetOrCreateAssessmentState(userID int, totalQuestions int) (*models.AssessmentState, error) {
	// First, try to insert a new assessment.
	// ON CONFLICT(user_id) WHERE is_complete = false DO NOTHING
	// This ensures that we only insert if there is no other in-progress assessment for this user.
	order := make([]int, totalQuestions)
	for i := range order {
		order[i] = i
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	order64 := make([]int64, len(order))
	for i, v := range order {
		order64[i] = int64(v)
	}

	insertQuery := `
		INSERT INTO assessments (user_id, question_order, current_question_index)
		VALUES ($1, $2, 0)
		ON CONFLICT(user_id) WHERE is_complete = false
		DO NOTHING
	`
	_, err := database.DB.Exec(insertQuery, userID, pq.Array(order64))
	if err != nil {
		return nil, err
	}

	// Now, whether we inserted or not, select the active assessment.
	// This will return either the one we just created or the one that already existed.
	state := &models.AssessmentState{}
	selectQuery := `
		SELECT id, user_id, is_complete, question_order, current_question_index
		FROM assessments
		WHERE user_id = $1 AND is_complete = false
		LIMIT 1
	`
	err = database.DB.QueryRow(selectQuery, userID).Scan(&state.ID, &state.UserID, &state.IsComplete, &state.QuestionOrder, &state.CurrentQuestionIndex)
	return state, err
}

func SaveAnswerAndUpdateState(assessmentID int, questionID string, answer string, nextIndex int) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	answerQuery := `INSERT INTO answers (assessment_id, question_id, answer_value) VALUES ($1, $2, $3)`
	_, err = tx.Exec(answerQuery, assessmentID, questionID, answer)
	if err != nil {
		return err
	}

	stateQuery := `UPDATE assessments SET current_question_index = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = tx.Exec(stateQuery, nextIndex, assessmentID)
	if err != nil {
		return err
	}

	return tx.Commit()
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
