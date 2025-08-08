// server/internal/repository/charts.go
package repository

import (
	"context"
	"crapp-go/internal/database"
	"fmt"
	"time"
)

type TimelineDataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type CorrelationDataPoint struct {
	MetricValue  float64 `json:"metricValue"`
	SymptomValue float64 `json:"symptomValue"`
}

func getMetricsCTE() string {
	return `
	WITH all_metrics AS (
		-- Mouse and Keyboard Metrics
		SELECT
			m.assessment_id,
			a.created_at,
			m.question_id,
			m.metric_key,
			m.metric_value
		FROM assessment_metrics m
		JOIN assessment_states a ON m.assessment_id = a.id
		
		UNION ALL
		
		-- Self-Reported Symptom Scores
		SELECT 
			ans.assessment_id, 
			a.created_at, 
			ans.question_id, 
			ans.question_id as metric_key, -- For symptoms, the metric_key is the question_id
			CASE
				WHEN ans.answer_value ~ '^[0-9\.]+$' THEN ans.answer_value::float
				ELSE NULL
			END as metric_value
		FROM answers ans
		JOIN assessment_states a ON ans.assessment_id = a.id
		WHERE ans.question_id IN ('headache', 'cognitive', 'tinnitus', 'dizziness', 'visual', 'medication_changes')

		UNION ALL

		-- CPT Results
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'reaction_time' AS metric_key, average_reaction_time AS metric_value FROM cpt_results UNION ALL
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'detection_rate' AS metric_key, detection_rate AS metric_value FROM cpt_results UNION ALL
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'omission_error_rate' AS metric_key, omission_error_rate AS metric_value FROM cpt_results UNION ALL
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'commission_error_rate' AS metric_key, commission_error_rate AS metric_value FROM cpt_results

		UNION ALL

		-- TMT Results
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_a_time' AS metric_key, part_a_completion_time AS metric_value FROM tmt_results UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_b_time' AS metric_key, part_b_completion_time AS metric_value FROM tmt_results UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_a_errors' AS metric_key, part_a_errors::float AS metric_value FROM tmt_results UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_b_errors' AS metric_key, part_b_errors::float AS metric_value FROM tmt_results UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'b_a_ratio' AS metric_key, b_to_a_ratio AS metric_value FROM tmt_results

		UNION ALL

		-- DST Results
		SELECT assessment_id, created_at, 'dst' AS question_id, 'highest_span' AS metric_key, highest_span_achieved::float AS metric_value FROM dst_results UNION ALL
		SELECT assessment_id, created_at, 'dst' AS question_id, 'correct_trials' AS metric_key, correct_trials::float AS metric_value FROM dst_results UNION ALL
		SELECT assessment_id, created_at, 'dst' AS question_id, 'total_trials' AS metric_key, total_trials::float AS metric_value FROM dst_results
	)
	`
}

func GetTimelineData(ctx context.Context, userID int, taskID string, metricKey string) ([]TimelineDataPoint, error) {
	var data []TimelineDataPoint

	query := fmt.Sprintf(`
		%s
		SELECT
			am.created_at as date,
			am.metric_value as value
		FROM all_metrics am
		JOIN assessment_states a ON am.assessment_id = a.id
		WHERE a.user_id = ? AND am.question_id = ? AND am.metric_key = ? AND a.is_complete = true
		ORDER BY am.created_at;
	`, getMetricsCTE())

	err := database.DB.WithContext(ctx).Raw(query, userID, taskID, metricKey).Scan(&data).Error

	return data, err
}

func GetCorrelationData(ctx context.Context, userID int, symptomQuestionID, taskID, metricKey string) ([]CorrelationDataPoint, error) {
	var data []CorrelationDataPoint
	query := fmt.Sprintf(`
		%s
		SELECT
			task_metric.metric_value AS metric_value,
			symptom.metric_value AS symptom_value
		FROM
			(
				SELECT assessment_id, metric_value
				FROM all_metrics
				WHERE question_id = ? AND metric_key = ?
			) AS task_metric
		JOIN
			(
				SELECT assessment_id, metric_value
				FROM all_metrics
				WHERE question_id = ? AND metric_key = ?
			) AS symptom ON task_metric.assessment_id = symptom.assessment_id
		JOIN assessment_states a ON task_metric.assessment_id = a.id
		WHERE a.user_id = ? AND a.is_complete = true;
	`, getMetricsCTE())

	err := database.DB.WithContext(ctx).Raw(query, taskID, metricKey, symptomQuestionID, symptomQuestionID, userID).Scan(&data).Error
	return data, err
}
