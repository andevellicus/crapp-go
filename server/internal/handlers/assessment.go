package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"crapp-go/internal/metrics"
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

// Start begins or resumes an assessment for the logged-in user.
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

	// If the assessment is already complete, show the results.
	if state.CurrentQuestionIndex >= len(state.QuestionOrder) {
		h.showResults(c, state.ID)
		return
	}

	currentQuestion := h.Assessment.Questions[state.QuestionOrder[state.CurrentQuestionIndex]]

	// Prepare settings JSON in the handler.
	settingsJSON := h.prepareSettingsJSON(currentQuestion)

	csrfToken, exists := c.Get("csrf_token")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	cspNonce, exists := c.Get("csp_nonce")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	component := views.AssessmentPage(currentQuestion, state.CurrentQuestionIndex, len(state.QuestionOrder), "", settingsJSON, csrfToken.(string), cspNonce.(string))

	if isHTMX {
		component.Render(c.Request.Context(), c.Writer)
	} else {
		views.Layout("Assessment", true, csrfToken.(string), cspNonce.(string)).Render(templ.WithChildren(c.Request.Context(), component), c.Writer)
	}
}

// NextQuestion saves the answer for the current question and serves the next one.
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

	currentQuestion := h.Assessment.Questions[state.QuestionOrder[state.CurrentQuestionIndex]]
	questionID := c.PostForm("questionId")
	answer := c.PostForm("answer")

	switch currentQuestion.Type {
	case "cpt":
		if answer != "" {
			var data metrics.CPTData
			if err := json.Unmarshal([]byte(answer), &data); err == nil {
				summary, events := processCPTData(&data, state.ID)
				if err := repository.SaveCPTResultTx(summary, events); err != nil {
					h.log.Error("Failed to save CPT transaction", zap.Error(err), zap.Int("assessmentID", state.ID))
				}
			} else {
				h.log.Error("Failed to unmarshal CPT data", zap.Error(err))
			}
		}

	case "dst":
		if answer != "" {
			var data metrics.DigitSpanRawData
			if err := json.Unmarshal([]byte(answer), &data); err == nil {
				summary := processDSTData(&data, state.ID)
				if err := repository.SaveDSTResultTx(summary, data.Results); err != nil {
					h.log.Error("Failed to save DST transaction", zap.Error(err), zap.Int("assessmentID", state.ID))
				}
			} else {
				h.log.Error("Failed to unmarshal DST data", zap.Error(err))
			}
		}

	case "tmt":
		if answer != "" {
			var data metrics.TrailMakingData
			if err := json.Unmarshal([]byte(answer), &data); err == nil {
				summary := processTMTData(&data, state.ID)
				if err := repository.SaveTMTResultTx(summary, data.Clicks); err != nil {
					h.log.Error("Failed to save TMT transaction", zap.Error(err), zap.Int("assessmentID", state.ID))
				}
			} else {
				h.log.Error("Failed to unmarshal TMT data", zap.Error(err))
			}
		}

	default:
		if currentQuestion.Required && answer == "" {
			errorMessage := "This question is required. Please select an answer."

			// Get the CSRF token to pass back to the template
			csrfToken, exists := c.Get("csrf_token")
			if !exists {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			cspNonce, exists := c.Get("csp_nonce")
			if !exists {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			// Re-render the same page with an error message AND the CSRF token
			views.AssessmentPage(currentQuestion, state.CurrentQuestionIndex, len(state.QuestionOrder), errorMessage, "", csrfToken.(string), cspNonce.(string)).Render(c, c.Writer)
			return // Stop processing
		}
		if err := repository.SaveAnswer(state.ID, questionID, answer); err != nil {
			h.log.Error("Could not save answer", zap.Error(err), zap.Int("assessmentID", state.ID))
			c.String(http.StatusInternalServerError, "Could not save answer")
			return
		}
	}

	// --- Advance to the next state ---
	nextIndex := state.CurrentQuestionIndex + 1
	repository.UpdateAssessmentIndex(state.ID, nextIndex)

	if nextIndex >= len(state.QuestionOrder) {
		repository.CompleteAssessment(state.ID)
		h.showResults(c, state.ID)
	} else {
		nextQuestion := h.Assessment.Questions[state.QuestionOrder[nextIndex]]
		settingsJSON := h.prepareSettingsJSON(nextQuestion)
		csrfToken, exists := c.Get("csrf_token")
		if !exists {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		cspNonce, exists := c.Get("csp_nonce")
		if !exists {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		views.AssessmentPage(nextQuestion, nextIndex, len(state.QuestionOrder), "", settingsJSON, csrfToken.(string), cspNonce.(string)).Render(c, c.Writer)
	}
}

// PreviousQuestion moves the assessment state to the previous question.
func (h *AssessmentHandler) PreviousQuestion(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(int)
	if !ok {
		c.Redirect(http.StatusFound, "/")
		return
	}

	state, err := repository.GetOrCreateAssessmentState(userID, len(h.Assessment.Questions))
	if err != nil {
		h.log.Error("Could not get assessment state for prev", zap.Error(err), zap.Int("userID", userID))
		c.String(http.StatusInternalServerError, "Could not get assessment state")
		return
	}

	prevIndex := state.CurrentQuestionIndex - 1
	if prevIndex < 0 {
		prevIndex = 0 // Safeguard
	}

	if err := repository.UpdateAssessmentIndex(state.ID, prevIndex); err != nil {
		h.log.Error("Could not update state for prev", zap.Error(err), zap.Int("assessmentID", state.ID))
		c.String(http.StatusInternalServerError, "Could not update state")
		return
	}

	prevQuestion := h.Assessment.Questions[state.QuestionOrder[prevIndex]]
	settingsJSON := h.prepareSettingsJSON(prevQuestion)
	csrfToken, exists := c.Get("csrf_token")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	cspNonce, exists := c.Get("csp_nonce")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	views.AssessmentPage(prevQuestion, prevIndex, len(state.QuestionOrder), "", settingsJSON, csrfToken.(string), cspNonce.(string)).Render(c, c.Writer)
}

func (h *AssessmentHandler) prepareSettingsJSON(question models.Question) string {
	if question.Type != "cpt" && question.Type != "dst" && question.Type != "tmt" {
		return "" // No settings needed for standard questions
	}
	settings := make(map[string]interface{})
	for _, option := range question.Options {
		settings[option.Label] = option.Value
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		h.log.Error("Failed to marshal question settings", zap.Error(err), zap.String("questionId", question.ID))
		return "{}" // Return empty JSON object on error
	}
	return string(settingsJSON)
}

// showResults renders the final assessment results page.
func (h *AssessmentHandler) showResults(c *gin.Context, assessmentID int) {
	answers, err := repository.GetAnswersForAssessment(assessmentID)
	if err != nil {
		h.log.Error("Error getting answers for results page", zap.Error(err), zap.Int("assessmentID", assessmentID))
		c.String(http.StatusInternalServerError, "Could not load assessment results.")
		return
	}
	views.AssessmentResults(answers).Render(c, c.Writer)
}

// --- Data Processing Helpers ---

func processCPTData(data *metrics.CPTData, assessmentID int) (models.CPTResult, []models.CPTEvent) {
	summary := models.CPTResult{
		AssessmentID:        assessmentID,
		CorrectDetections:   metrics.CountCorrectDetections(data),
		CommissionErrors:    metrics.CountCommissionErrors(data),
		OmissionErrors:      metrics.CountOmissionErrors(data),
		AverageReactionTime: metrics.CalculateAverageReactionTime(data),
		ReactionTimeSD:      metrics.CalculateReactionTimeSD(data),
		DetectionRate:       metrics.CalculateDetectionRate(data),
		OmissionErrorRate:   metrics.CalculateOmissionErrorRate(data),
		CommissionErrorRate: metrics.CalculateCommissionErrorRate(data),
		CreatedAt:           time.Now(),
	}

	var events []models.CPTEvent
	for _, stim := range data.StimuliPresented {
		s := stim // Local copy for pointer safety
		events = append(events, models.CPTEvent{
			EventType:     "stimulus",
			StimulusValue: &s.Value,
			IsTarget:      &s.IsTarget,
			PresentedAt:   &s.PresentedAt,
		})
	}
	for _, resp := range data.Responses {
		r := resp // Local copy for pointer safety
		events = append(events, models.CPTEvent{
			EventType:     "response",
			StimulusValue: &r.Stimulus,
			IsTarget:      &r.IsTarget,
			ResponseTime:  &r.ResponseTime,
			StimulusIndex: &r.StimulusIndex,
		})
	}

	return summary, events
}

func processDSTData(data *metrics.DigitSpanRawData, assessmentID int) models.DSTResult {
	processedResult, _ := metrics.CalculateDigitSpanMetrics(data)
	return models.DSTResult{
		AssessmentID:        assessmentID,
		HighestSpanAchieved: processedResult.HighestSpanAchieved,
		TotalTrials:         processedResult.TotalTrials,
		CorrectTrials:       processedResult.CorrectTrials,
		CreatedAt:           time.Now(),
	}
}

func processTMTData(data *metrics.TrailMakingData, assessmentID int) models.TMTResult {
	processedResult := metrics.CalculateTrailMetrics(data)
	return models.TMTResult{
		AssessmentID:        assessmentID,
		PartACompletionTime: processedResult.PartACompletionTime,
		PartAErrors:         processedResult.PartAErrors,
		PartBCompletionTime: processedResult.PartBCompletionTime,
		PartBErrors:         processedResult.PartBErrors,
		BToARatio:           processedResult.BToARatio,
		CreatedAt:           time.Now(),
	}
}
