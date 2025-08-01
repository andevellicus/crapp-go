package handlers

import (
	"net/http"

	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/views"

	"github.com/a-h/templ"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AssessmentHandler struct {
	log        *zap.Logger
	Assessment *models.Assessment
}

func NewAssessmentHandler(log *zap.Logger, assessment *models.Assessment) *AssessmentHandler {
	return &AssessmentHandler{log: log, Assessment: assessment}
}

func (h *AssessmentHandler) Start(c *gin.Context, isHTMX bool) {
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(int)
	if !ok {
		c.Redirect(http.StatusFound, "/")
		return
	}

	state, err := repository.GetOrCreateAssessmentState(userID, len(h.Assessment.Questions))
	if err != nil {
		h.log.Error("Error getting assessment state", zap.Error(err), zap.Int("userID", userID))
		c.String(http.StatusInternalServerError, "Could not start or resume assessment")
		return
	}

	// Get the actual question using the index from the database
	questionIndex := state.QuestionOrder[state.CurrentQuestionIndex]
	currentQuestion := h.Assessment.Questions[questionIndex]

	component := views.AssessmentPage(currentQuestion, state.CurrentQuestionIndex, len(state.QuestionOrder), "")

	if isHTMX {
		component.Render(c.Request.Context(), c.Writer)
	} else {
		// Render the full page for direct navigation
		views.Layout("Assessment", true).Render(templ.WithChildren(c.Request.Context(), component), c.Writer)
	}
}

func (h *AssessmentHandler) PreviousQuestion(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(int)
	if !ok {
		c.Redirect(http.StatusFound, "/")
		return
	}

	state, err := repository.GetOrCreateAssessmentState(userID, len(h.Assessment.Questions))
	if err != nil {
		h.log.Error("Could not get assessment state", zap.Error(err), zap.Int("userID", userID))
		c.String(http.StatusInternalServerError, "Could not get assessment state")
		return
	}

	prevIndex := state.CurrentQuestionIndex - 1
	if prevIndex < 0 {
		prevIndex = 0 // Should not happen if button is hidden, but as a safeguard
	}

	// Only need to update the index in the database
	err = repository.UpdateAssessmentIndex(state.ID, prevIndex)
	if err != nil {
		h.log.Error("Could not update state", zap.Error(err), zap.Int("assessmentID", state.ID))
		c.String(http.StatusInternalServerError, "Could not update state")
		return
	}

	prevQuestionIndex := state.QuestionOrder[prevIndex]
	prevQuestion := h.Assessment.Questions[prevQuestionIndex]

	views.AssessmentPage(prevQuestion, prevIndex, len(state.QuestionOrder), "").Render(c, c.Writer)
}

func (h *AssessmentHandler) NextQuestion(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(int)
	if !ok {
		c.Redirect(http.StatusFound, "/")
		return
	}

	state, err := repository.GetOrCreateAssessmentState(userID, len(h.Assessment.Questions))
	if err != nil {
		h.log.Error("Could not get assessment state", zap.Error(err), zap.Int("userID", userID))
		c.String(http.StatusInternalServerError, "Could not get assessment state")
		return
	}

	// Get the current question to check its properties
	currentQuestionIndexInOrder := state.QuestionOrder[state.CurrentQuestionIndex]
	currentQuestion := h.Assessment.Questions[currentQuestionIndexInOrder]

	questionID := c.PostForm("questionId")
	answer := c.PostForm("answer")

	if currentQuestion.Required && answer == "" {
		errorMessage := "This question is required. Please select an answer."
		// Re-render the same page with an error message
		views.AssessmentPage(currentQuestion, state.CurrentQuestionIndex, len(state.QuestionOrder), errorMessage).Render(c, c.Writer)
		return // Stop processing
	}

	nextIndex := state.CurrentQuestionIndex + 1

	err = repository.SaveAnswerAndUpdateState(state.ID, questionID, answer, nextIndex)
	if err != nil {
		h.log.Error("Could not save answer", zap.Error(err), zap.Int("assessmentID", state.ID))
		c.String(http.StatusInternalServerError, "Could not save answer")
		return
	}

	if nextIndex >= len(state.QuestionOrder) {
		repository.CompleteAssessment(state.ID)
		answers, _ := repository.GetAnswersForAssessment(state.ID)
		err = views.AssessmentResults(answers).Render(c, c.Writer)
		if err != nil {
			h.log.Error("Error rendering assessment results", zap.Error(err), zap.Int("assessmentID", state.ID))
		}
		h.log.Info("Assessment completed", zap.Int("assessmentID", state.ID))
		return
	}

	nextQuestionIndexInOrder := state.QuestionOrder[nextIndex]
	nextQuestion := h.Assessment.Questions[nextQuestionIndexInOrder]

	// Pass an empty string for the error message on success
	views.AssessmentPage(nextQuestion, nextIndex, len(state.QuestionOrder), "").Render(c, c.Writer)
}
