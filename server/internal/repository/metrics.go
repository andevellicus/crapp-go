// server/internal/repository/metrics.go
package repository

import (
	"crapp-go/internal/database"
	"crapp-go/internal/models"
)

func SaveMetric(metric models.AssessmentMetric) error {
	return database.DB.Create(&metric).Error
}
