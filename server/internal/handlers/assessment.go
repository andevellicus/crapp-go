package handlers

import (
	"log"
	"net/http"

	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/views"

	"github.com/a-h/templ"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AssessmentHandler struct {
	Assessment *models.Assessment
}

func NewAssessmentHandler(assessment *models.Assessment) *AssessmentHandler {
	return &AssessmentHandler{Assessment: assessment}
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
		log.Printf("Error getting assessment state: %v", err)
		c.String(http.StatusInternalServerError, "Could not start or resume assessment")
		return
	}

	// Get the actual question using the index from the database
	questionIndex := state.QuestionOrder[state.CurrentQuestionIndex]
	currentQuestion := h.Assessment.Questions[questionIndex]

	component := views.AssessmentPage(currentQuestion, state.CurrentQuestionIndex, len(state.QuestionOrder))

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
		c.String(http.StatusInternalServerError, "Could not update state")
		return
	}

	prevQuestionIndex := state.QuestionOrder[prevIndex]
	prevQuestion := h.Assessment.Questions[prevQuestionIndex]

	views.AssessmentPage(prevQuestion, prevIndex, len(state.QuestionOrder)).Render(c, c.Writer)
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
		c.String(http.StatusInternalServerError, "Could not get assessment state")
		return
	}

	questionID := c.PostForm("questionId")
	answer := c.PostForm("answer")
	nextIndex := state.CurrentQuestionIndex + 1

	err = repository.SaveAnswerAndUpdateState(state.ID, questionID, answer, nextIndex)
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not save answer")
		return
	}

	if nextIndex >= len(state.QuestionOrder) {
		repository.CompleteAssessment(state.ID)
		answers, _ := repository.GetAnswersForAssessment(state.ID)
		views.AssessmentResults(answers).Render(c, c.Writer)
		return
	}

	// Get the next question using the updated index
	nextQuestionIndex := state.QuestionOrder[nextIndex]
	nextQuestion := h.Assessment.Questions[nextQuestionIndex]

	views.AssessmentPage(nextQuestion, nextIndex, len(state.QuestionOrder)).Render(c, c.Writer)
}
