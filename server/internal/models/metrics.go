// server/internal/models/metrics.go
package models

import "time"

type MetricResult struct {
	Value      float64 `json:"value"`
	Calculated bool    `json:"calculated"`
	SampleSize int     `json:"sampleSize,omitempty"`
}

type AssessmentMetric struct {
	ID           int
	AssessmentID int
	QuestionID   string
	MetricKey    string
	MetricValue  float64
	SampleSize   int
	CreatedAt    time.Time
}

type InteractionData struct {
	MouseMovements    []MouseMovement    `json:"movements"`
	MouseInteractions []MouseInteraction `json:"interactions"`
	KeyboardEvents    []KeyboardEvent    `json:"keyboardEvents"`
	StartTime         float64            `json:"startTime"`
}

type MouseMovement struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Timestamp  float64 `json:"timestamp"`
	TargetID   string  `json:"targetId,omitempty"`
	QuestionID string  `json:"questionId,omitempty"`
}

type MouseInteraction struct {
	TargetID   string  `json:"targetId"`
	TargetType string  `json:"targetType"`
	QuestionID string  `json:"questionId,omitempty"`
	ClickX     float64 `json:"clickX"`
	ClickY     float64 `json:"clickY"`
	TargetX    float64 `json:"targetX"`
	TargetY    float64 `json:"targetY"`
	Timestamp  float64 `json:"timestamp"`
}

type KeyboardEvent struct {
	Type       string  `json:"type"`
	Key        string  `json:"key"`
	IsModifier bool    `json:"isModifier"`
	Timestamp  float64 `json:"timestamp"`
	QuestionID string  `json:"questionId,omitempty"`
}
