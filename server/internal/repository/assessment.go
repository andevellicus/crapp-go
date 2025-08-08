package repository

import (
	"math/rand"
	"time"

	"crapp-go/internal/database"
	"crapp-go/internal/models"

	"gorm.io/gorm"
)

func GetOrCreateAssessmentState(userID uint, totalQuestions int) (*models.AssessmentState, error) {
	var state models.AssessmentState
	// Attempt to find an incomplete assessment
	err := database.DB.Where("user_id = ? AND is_complete = ?", userID, false).First(&state).Error

	// Case 1: An active assessment was found successfully.
	if err == nil {
		return &state, nil
	}

	// Case 2: No record was found. This is the expected path for a new assessment.
	if err == gorm.ErrRecordNotFound {
		// If we're not meant to create, just return the not found error.
		if totalQuestions <= 0 {
			return nil, err
		}

		// Proceed to create a new one.
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

		newState := models.AssessmentState{
			UserID:               int(userID),
			QuestionOrder:        order64,
			CurrentQuestionIndex: 0,
			IsComplete:           false,
		}
		// Attempt to create the new state and return any errors from that operation.
		if createErr := database.DB.Create(&newState).Error; createErr != nil {
			return nil, createErr
		}
		return &newState, nil
	}

	// Case 3: A different, unexpected database error occurred. Return it.
	return nil, err
}

// GetMostRecentAssessmentState finds the most recently updated assessment for a user, regardless of completion status.
func GetMostRecentAssessmentState(userID uint) (*models.AssessmentState, error) {
	var state models.AssessmentState
	err := database.DB.Where("user_id = ?", userID).Order("updated_at desc").First(&state).Error
	return &state, err
}

func UpdateAssessmentIndex(assessmentID uint, newIndex int) error {
	return database.DB.Model(&models.AssessmentState{}).Where("id = ?", assessmentID).Update("current_question_index", newIndex).Error
}

func CompleteAssessment(assessmentID uint) error {
	return database.DB.Model(&models.AssessmentState{}).Where("id = ?", assessmentID).Update("is_complete", true).Error
}

func GetAnswersForAssessment(assessmentID uint) (map[string]string, error) {
	var answers []models.Answer
	if err := database.DB.Where("assessment_id = ?", assessmentID).Find(&answers).Error; err != nil {
		return nil, err
	}

	answerMap := make(map[string]string)
	for _, answer := range answers {
		answerMap[answer.QuestionID] = answer.AnswerValue
	}
	return answerMap, nil
}
