// server/internal/handlers/results.go
package handlers

import (
	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/views"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type ResultsHandler struct {
	log        *zap.Logger
	Assessment *models.Assessment
}

func NewResultsHandler(log *zap.Logger, assessment *models.Assessment) *ResultsHandler {
	return &ResultsHandler{log: log, Assessment: assessment}
}

func (h *ResultsHandler) ShowResults(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(int)
	if !ok {
		c.Redirect(http.StatusFound, "/")
		return
	}

	assessment := h.Assessment
	primaryTaskID := c.Query("symptom") // Renamed for clarity in the template, but it's the task/question ID
	metricKey := c.Query("metric")

	// Group questions by their function for the dropdown
	questionGroups := make(map[string][]models.Question)
	for _, q := range assessment.Questions {
		var groupKey string
		switch q.Type {
		// Scales should use the radio type, otherwise use a drop down
		case "radio":
			groupKey = "symptom"
		case "cpt", "tmt", "dst":
			groupKey = q.Type
		default:
			groupKey = q.MetricsType
		}
		questionGroups[groupKey] = append(questionGroups[groupKey], q)
	}

	if primaryTaskID == "" {
		if len(questionGroups["symptom"]) > 0 {
			primaryTaskID = questionGroups["symptom"][0].ID
		} else if len(assessment.Questions) > 0 {
			primaryTaskID = assessment.Questions[0].ID
		}
	}

	selectedQuestion, questionFound := getQuestionByID(primaryTaskID, assessment.Questions)
	if !questionFound {
		c.String(http.StatusBadRequest, "Invalid question selected")
		return
	}

	availableMetrics := getAvailableMetrics(selectedQuestion)
	metricLabel := strings.Title(strings.ReplaceAll(metricKey, "_", " "))
	// Check if the current metricKey is valid for the selected question
	isMetricValid := false
	for _, metric := range availableMetrics {
		if metric.Value == metricKey {
			isMetricValid = true
			metricLabel = metric.Label
			break
		}
	}

	// If the metric is not valid, reset it to the first available metric
	if !isMetricValid && len(availableMetrics) > 0 {
		metricKey = availableMetrics[0].Value
		metricLabel = availableMetrics[0].Label
	}

	var correlationSymptomID string
	var correlationData []repository.CorrelationDataPoint
	showCorrelationChart := false

	// Correlation is only shown if the selected item is a task, not a symptom report.
	if selectedQuestion.Type == "radio" {
		if len(questionGroups["symptom"]) > 0 {
			showCorrelationChart = true
			correlationSymptomID = questionGroups["symptom"][0].ID // Correlate against the first symptom
		}
	}

	// Fetch data for the timeline chart.
	timelineData, err := repository.GetTimelineData(c, userID, primaryTaskID, metricKey)
	if err != nil {
		h.log.Error("Failed to get timeline data", zap.Error(err), zap.String("taskID", primaryTaskID), zap.String("metricKey", metricKey))
		c.String(http.StatusInternalServerError, "Failed to load timeline data")
		return
	}

	// Fetch data for the correlation chart if needed.
	if showCorrelationChart {
		var err error
		correlationData, err = repository.GetCorrelationData(c, userID, correlationSymptomID, primaryTaskID, metricKey)
		if err != nil {
			h.log.Error("Failed to get correlation data", zap.Error(err), zap.String("symptomID", correlationSymptomID), zap.String("taskID", primaryTaskID), zap.String("metricKey", metricKey))
			c.String(http.StatusInternalServerError, "Failed to load correlation data")
			return
		}
	}

	timelineChart := generateTimelineChart(timelineData, metricLabel)
	correlationChart := generateCorrelationChart(correlationData, metricLabel, correlationSymptomID)

	timelineOptionsJSON, _ := json.Marshal(timelineChart.JSON())
	correlationOptionsJSON, _ := json.Marshal(correlationChart.JSON())

	csrfToken, _ := c.Get("csrf_token")
	cspNonce, _ := c.Get("csp_nonce")

	metricsTypeForExplanation := selectedQuestion.Type
	if metricsTypeForExplanation != "cpt" && metricsTypeForExplanation != "tmt" && metricsTypeForExplanation != "dst" {
		metricsTypeForExplanation = selectedQuestion.MetricsType
	}

	component := views.ResultsCharts(
		questionGroups,
		availableMetrics,
		primaryTaskID,
		metricKey,
		string(timelineOptionsJSON),
		string(correlationOptionsJSON),
		cspNonce.(string),
		metricsTypeForExplanation,
		showCorrelationChart,
	)

	if c.GetHeader("HX-Request") == "true" {
		component.Render(c.Request.Context(), c.Writer)
	} else {
		views.Layout("Results", true, csrfToken.(string), cspNonce.(string)).Render(
			templ.WithChildren(c.Request.Context(), component),
			c.Writer,
		)
	}
}

// getAvailableMetrics now correctly combines metrics from the Question's TYPE and its Metrics TYPE.
func getAvailableMetrics(question models.Question) []models.MetricOption {
	// For cognitive tests, use the test type
	switch question.Type {
	case "cpt":
		return []models.MetricOption{
			{Value: "reaction_time", Label: "Reaction Time (ms)"},
			{Value: "detection_rate", Label: "Detection Rate (%)"},
			{Value: "omission_error_rate", Label: "Omission Error Rate (%)"},
			{Value: "commission_error_rate", Label: "Commission Error Rate (%)"},
		}
	case "tmt":
		return []models.MetricOption{
			{Value: "part_a_time", Label: "Part A Time (ms)"},
			{Value: "part_b_time", Label: "Part B Time (ms)"},
			{Value: "b_a_ratio", Label: "B/A Ratio"},
			{Value: "part_a_errors", Label: "Part A Errors"},
			{Value: "part_b_errors", Label: "Part B Errors"},
		}
	case "dst":
		return []models.MetricOption{
			{Value: "highest_span", Label: "Highest Span Achieved"},
			{Value: "correct_trials", Label: "Correct Trials"},
			{Value: "total_trials", Label: "Total Trials"},
		}
	default:
		// For regular questions, use the metrics_type
		switch question.MetricsType {
		case "keyboard":
			return []models.MetricOption{
				{Value: "typing_speed", Label: "Typing Speed"},
				{Value: "average_inter_key_interval", Label: "Inter-Key Interval"},
				{Value: "typing_rhythm_variability", Label: "Typing Rhythm Variability"},
				{Value: "correction_rate", Label: "Correction Rate"},
				{Value: "keyboard_fluency", Label: "Keyboard Fluency Score"},
			}
		case "mouse":
			return []models.MetricOption{
				{Value: "click_precision", Label: "Click Precision"},
				{Value: "path_efficiency", Label: "Path Efficiency"},
				{Value: "overshoot_rate", Label: "Overshoot Rate"},
				{Value: "average_velocity", Label: "Average Velocity"},
				{Value: "velocity_variability", Label: "Velocity Variability"},
			}
		default:
			// Fallback to mouse metrics
			return []models.MetricOption{
				{Value: "click_precision", Label: "Click Precision"},
				{Value: "path_efficiency", Label: "Path Efficiency"},
				{Value: "overshoot_rate", Label: "Overshoot Rate"},
				{Value: "average_velocity", Label: "Average Velocity"},
				{Value: "velocity_variability", Label: "Velocity Variability"},
			}
		}
	}
}

func getQuestionByID(id string, questions []models.Question) (models.Question, bool) {
	for _, q := range questions {
		if q.ID == id {
			return q, true
		}
	}
	return models.Question{}, false
}

func generateTimelineChart(data []repository.TimelineDataPoint, metricLabel string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Metric Over Time",
			Subtitle: metricLabel,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "time", // Change type from "category" to "time"
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
			// Optional: helps the axis scale nicely
			Scale: opts.Bool(true),
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider"}),
	)

	// Create data points in the format [date, value]
	items := make([]opts.LineData, 0)
	for _, point := range data {
		items = append(items, opts.LineData{Value: []interface{}{point.Date, point.Value}})
	}

	line.AddSeries(metricLabel, items).SetSeriesOptions(charts.WithLineStyleOpts(opts.LineStyle{Width: 2}))
	return line
}

func generateCorrelationChart(data []repository.CorrelationDataPoint, metricKey, symptomKey string) *charts.Scatter {
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Metric vs. Symptom Correlation",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "value",
			Name: strings.ReplaceAll(metricKey, "_", " "),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
			Name: strings.ReplaceAll(symptomKey, "_", " "),
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
	)

	items := make([]opts.ScatterData, 0)
	for _, point := range data {
		items = append(items, opts.ScatterData{Value: []interface{}{point.MetricValue, point.SymptomValue}})
	}

	scatter.AddSeries("Correlation", items)
	return scatter
}
