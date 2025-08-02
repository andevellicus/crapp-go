// server/internal/repository/metrics.go
package repository

import (
	"crapp-go/internal/database"
	"crapp-go/internal/models"
)

func SaveMetric(metric models.AssessmentMetric) error {
	query := `
		INSERT INTO metrics (assessment_id, question_id, metric_key, metric_value, sample_size)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := database.DB.Exec(query, metric.AssessmentID, metric.QuestionID, metric.MetricKey, metric.MetricValue, metric.SampleSize)
	return err
}
