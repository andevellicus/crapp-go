// server/internal/repository/charts.go
package repository

import (
	"context"
	"crapp-go/internal/database"
	"fmt"
	"time"
)

type TimelineDataPoint struct {
	Date  time.Time
	Value float64
}

type CorrelationDataPoint struct {
	MetricValue  float64
	SymptomValue float64
}

func getMetricsCTE() string {
	return `
	WITH all_metrics AS (
		SELECT
			m.assessment_id,
			a.created_at,
			m.question_id,
			m.metric_key,
			m.metric_value
		FROM metrics m
		JOIN assessments a ON m.assessment_id = a.id
		UNION ALL
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'reaction_time', average_reaction_time FROM cpt_results
		UNION ALL
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'detection_rate', detection_rate FROM cpt_results
		UNION ALL
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'omission_error_rate', omission_error_rate FROM cpt_results
		UNION ALL
		SELECT assessment_id, created_at, 'cpt' AS question_id, 'commission_error_rate', commission_error_rate FROM cpt_results
		UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_a_time', part_a_completion_time FROM tmt_results
		UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_b_time', part_b_completion_time FROM tmt_results
		UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_a_errors', part_a_errors::float FROM tmt_results
		UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'part_b_errors', part_b_errors::float FROM tmt_results
		UNION ALL
		SELECT assessment_id, created_at, 'tmt' AS question_id, 'b_a_ratio', b_to_a_ratio FROM tmt_results
		UNION ALL
		SELECT assessment_id, created_at, 'dst' AS question_id, 'highest_span', highest_span_achieved::float FROM dst_results
		UNION ALL
		SELECT assessment_id, created_at, 'dst' AS question_id, 'correct_trials', correct_trials::float FROM dst_results
		UNION ALL
		SELECT assessment_id, created_at, 'dst' AS question_id, 'total_trials', total_trials::float FROM dst_results
		UNION ALL
		SELECT ans.assessment_id, a.created_at, ans.question_id, ans.question_id as metric_key,
			CASE
				WHEN ans.answer_value ~ '^[0-9\.]+$' THEN ans.answer_value::float
				ELSE NULL
			END as metric_value
		FROM answers ans
		JOIN assessments a ON ans.assessment_id = a.id
	)
	`
}

func GetTimelineData(ctx context.Context, userID int, taskID string, metricKey string) ([]TimelineDataPoint, error) {
	query := fmt.Sprintf(`
		%s
		SELECT
			am.created_at,
			am.metric_value
		FROM all_metrics am
		JOIN assessments a ON am.assessment_id = a.id
		WHERE a.user_id = $1 AND am.question_id = $2 AND am.metric_key = $3 AND a.is_complete = true
		ORDER BY am.created_at;
	`, getMetricsCTE())

	rows, err := database.DB.QueryContext(ctx, query, userID, taskID, metricKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []TimelineDataPoint
	for rows.Next() {
		var point TimelineDataPoint
		if err := rows.Scan(&point.Date, &point.Value); err != nil {
			return nil, err
		}
		data = append(data, point)
	}
	return data, nil
}

func GetCorrelationData(ctx context.Context, userID int, symptomQuestionID, taskID, metricKey string) ([]CorrelationDataPoint, error) {
	query := fmt.Sprintf(`
		%s
		SELECT
			task_metric.metric_value AS metric_value,
			symptom.metric_value AS symptom_value
		FROM
			assessments a
		-- Join to get the interaction metric for the specific task
		JOIN
			all_metrics AS task_metric ON a.id = task_metric.assessment_id
			AND task_metric.question_id = $3 -- taskID
			AND task_metric.metric_key = $4 -- metricKey
		-- Join to get the symptom score for the same assessment
		JOIN
			all_metrics AS symptom ON a.id = symptom.assessment_id
			AND symptom.question_id = $2 -- symptomQuestionID
			AND symptom.metric_key = $2 -- for answers, metric_key is the same as question_id
		WHERE
			a.user_id = $1
			AND a.is_complete = true
			AND task_metric.metric_value IS NOT NULL
			AND symptom.metric_value IS NOT NULL;
	`, getMetricsCTE())

	rows, err := database.DB.QueryContext(ctx, query, userID, symptomQuestionID, taskID, metricKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []CorrelationDataPoint
	for rows.Next() {
		var point CorrelationDataPoint
		if err := rows.Scan(&point.MetricValue, &point.SymptomValue); err != nil {
			return nil, err
		}
		data = append(data, point)
	}
	return data, nil
}
