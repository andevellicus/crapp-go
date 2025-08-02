// server/internal/handlers/metrics.go
package handlers

import (
	"net/http"

	"crapp-go/internal/metrics"
	"crapp-go/internal/models"
	"crapp-go/internal/repository"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MetricsHandler struct {
	log *zap.Logger
}

func NewMetricsHandler(log *zap.Logger) *MetricsHandler {
	return &MetricsHandler{log: log}
}

func (h *MetricsHandler) SaveMetrics(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var interactionData models.InteractionData
	if err := c.ShouldBindJSON(&interactionData); err != nil {
		h.log.Error("Failed to bind interaction data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	calculatedMetrics := metrics.CalculateInteractionMetrics(&interactionData)

	state, err := repository.GetOrCreateAssessmentState(userID, 0) // Assuming we can get the state without question count
	if err != nil {
		h.log.Error("Failed to get assessment state", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assessment state"})
		return
	}

	for _, metric := range calculatedMetrics.GlobalMetrics {
		metric.AssessmentID = state.ID
		if err := repository.SaveMetric(metric); err != nil {
			h.log.Error("Failed to save global metric", zap.Error(err))
		}
	}

	for _, metric := range calculatedMetrics.QuestionMetrics {
		metric.AssessmentID = state.ID
		if err := repository.SaveMetric(metric); err != nil {
			h.log.Error("Failed to save question metric", zap.Error(err))
		}
	}

	c.Status(http.StatusOK)
}
