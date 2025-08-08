package metrics

import (
	"crapp-go/internal/models"
)

type MetricResult struct {
	Value      float64 `json:"value"`
	Calculated bool    `json:"calculated"`
	SampleSize int     `json:"sampleSize,omitempty"`
}

type CalculatedMetrics struct {
	GlobalMetrics   []models.AssessmentMetric
	QuestionMetrics []models.AssessmentMetric
}

// CalculateInteractionMetrics calculates all interaction metrics
func CalculateInteractionMetrics(interactions *models.InteractionData) *CalculatedMetrics {
	result := &CalculatedMetrics{
		GlobalMetrics:   []models.AssessmentMetric{},
		QuestionMetrics: []models.AssessmentMetric{},
	}

	globalInteractions := &models.InteractionData{}
	questionInteractions := make(map[string]*models.InteractionData)

	for _, m := range interactions.MouseMovements {
		if m.QuestionID == "" {
			globalInteractions.MouseMovements = append(globalInteractions.MouseMovements, m)
		} else {
			if _, ok := questionInteractions[m.QuestionID]; !ok {
				questionInteractions[m.QuestionID] = &models.InteractionData{}
			}
			questionInteractions[m.QuestionID].MouseMovements = append(questionInteractions[m.QuestionID].MouseMovements, m)
		}
	}

	for _, i := range interactions.MouseInteractions {
		if i.QuestionID == "" {
			globalInteractions.MouseInteractions = append(globalInteractions.MouseInteractions, i)
		} else {
			if _, ok := questionInteractions[i.QuestionID]; !ok {
				questionInteractions[i.QuestionID] = &models.InteractionData{}
			}
			questionInteractions[i.QuestionID].MouseInteractions = append(questionInteractions[i.QuestionID].MouseInteractions, i)
		}
	}

	for _, k := range interactions.KeyboardEvents {
		if k.QuestionID == "" {
			globalInteractions.KeyboardEvents = append(globalInteractions.KeyboardEvents, k)
		} else {
			if _, ok := questionInteractions[k.QuestionID]; !ok {
				questionInteractions[k.QuestionID] = &models.InteractionData{}
			}
			questionInteractions[k.QuestionID].KeyboardEvents = append(questionInteractions[k.QuestionID].KeyboardEvents, k)
		}
	}

	// --- Step 1: Calculate metrics ONLY for the global data ---
	globalMouseMetrics := map[string]MetricResult{
		"click_precision":      calculateClickPrecision(nil, globalInteractions),
		"path_efficiency":      calculatePathEfficiency(nil, globalInteractions),
		"overshoot_rate":       calculateOvershootRate(nil, globalInteractions),
		"average_velocity":     calculateAverageVelocity(nil, globalInteractions),
		"velocity_variability": calculateVelocityVariability(nil, globalInteractions),
	}
	globalKeyboardMetrics := calculateKeyboardMetrics(nil, globalInteractions)

	for key, val := range globalKeyboardMetrics {
		globalMouseMetrics[key] = val
	}

	for metricKey, metricResult := range globalMouseMetrics {
		if metricResult.Calculated {
			result.GlobalMetrics = append(result.GlobalMetrics, models.AssessmentMetric{
				QuestionID:  "global",
				MetricKey:   metricKey,
				MetricValue: metricResult.Value,
				SampleSize:  metricResult.SampleSize,
			})
		}
	}

	// --- Step 2: Calculate metrics for each question's bucket of data ---
	for questionID, specificInteractions := range questionInteractions {
		qID := questionID // Create a copy for the pointer

		qMouseMetrics := map[string]MetricResult{
			"click_precision":      calculateClickPrecision(&qID, specificInteractions),
			"path_efficiency":      calculatePathEfficiency(&qID, specificInteractions),
			"overshoot_rate":       calculateOvershootRate(&qID, specificInteractions),
			"average_velocity":     calculateAverageVelocity(&qID, specificInteractions),
			"velocity_variability": calculateVelocityVariability(&qID, specificInteractions),
		}
		qKeyboardMetrics := calculateKeyboardMetrics(&qID, specificInteractions)

		for key, val := range qKeyboardMetrics {
			qMouseMetrics[key] = val
		}

		for metricKey, metricResult := range qMouseMetrics {
			if metricResult.Calculated {
				result.QuestionMetrics = append(result.QuestionMetrics, models.AssessmentMetric{
					QuestionID:  questionID,
					MetricKey:   metricKey,
					MetricValue: metricResult.Value,
					SampleSize:  metricResult.SampleSize,
				})
			}
		}
	}

	return result
}
