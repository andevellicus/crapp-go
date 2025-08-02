package metrics

import (
	"math"
	"sort"

	"crapp-go/internal/models"
)

// calculateKeyboardMetrics calculates all keyboard-related metrics with enhanced analysis
func calculateKeyboardMetrics(questionID *string, interactions *models.InteractionData) map[string]MetricResult {
	events := filterKeyboardEventsByQuestion(questionID, interactions)
	metrics := make(map[string]MetricResult)

	// Initialize with uncalculated values
	metrics["typing_speed"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: len(events)}
	metrics["average_inter_key_interval"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["typing_rhythm_variability"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["average_key_hold_time"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["key_press_variability"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["correction_rate"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["immediate_correction_tendency"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["pause_rate"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["deep_thinking_pause_rate"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}
	metrics["keyboard_fluency"] = MetricResult{Value: 0.0, Calculated: false, SampleSize: 0}

	if len(events) < 3 {
		return metrics
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})

	keydownEvents := make([]models.KeyboardEvent, 0)
	for _, event := range events {
		if event.Type == "keydown" {
			keydownEvents = append(keydownEvents, event)
		}
	}

	if len(keydownEvents) >= 5 {
		contentKeys := 0
		for _, event := range keydownEvents {
			if len(event.Key) == 1 || event.Key == "Space" || event.Key == "Enter" {
				contentKeys++
			}
		}

		totalTime := (keydownEvents[len(keydownEvents)-1].Timestamp - keydownEvents[0].Timestamp) / 1000
		if totalTime > 0 && contentKeys > 0 {
			typingSpeed := float64(contentKeys) / totalTime
			metrics["typing_speed"] = MetricResult{Value: typingSpeed, Calculated: true, SampleSize: contentKeys}
		}
	}

	intervals := make([]float64, 0)
	for i := 1; i < len(keydownEvents); i++ {
		interval := keydownEvents[i].Timestamp - keydownEvents[i-1].Timestamp
		intervals = append(intervals, interval)
	}

	if len(intervals) >= 3 {
		sortedIntervals := make([]float64, len(intervals))
		copy(sortedIntervals, intervals)
		sort.Float64s(sortedIntervals)

		p95idx := int(float64(len(sortedIntervals)) * 0.95)
		if p95idx >= len(sortedIntervals) {
			p95idx = len(sortedIntervals) - 1
		}
		maxInterval := sortedIntervals[p95idx] * 1.5

		filteredIntervals := make([]float64, 0, len(intervals))
		for _, interval := range intervals {
			if interval <= maxInterval {
				filteredIntervals = append(filteredIntervals, interval)
			}
		}

		if len(filteredIntervals) >= 3 {
			var intervalSum float64
			for _, interval := range filteredIntervals {
				intervalSum += interval
			}
			avgInterval := intervalSum / float64(len(filteredIntervals))
			metrics["average_inter_key_interval"] = MetricResult{Value: avgInterval, Calculated: true, SampleSize: len(filteredIntervals)}

			var intervalVariance float64
			for _, interval := range filteredIntervals {
				intervalVariance += math.Pow(interval-avgInterval, 2)
			}
			intervalVariance /= float64(len(filteredIntervals) - 1)
			typingVariability := math.Sqrt(intervalVariance) / avgInterval
			metrics["typing_rhythm_variability"] = MetricResult{Value: typingVariability, Calculated: true, SampleSize: len(filteredIntervals)}
		}
	}

	if len(intervals) >= 5 {
		var avgInterval float64
		for _, interval := range intervals {
			avgInterval += interval
		}
		avgInterval /= float64(len(intervals))

		pauseThreshold := math.Max(avgInterval*3.0, 1000.0)
		pauseCount := 0
		longPauseCount := 0

		for _, interval := range intervals {
			if interval > pauseThreshold {
				pauseCount++
				if interval > 5000 {
					longPauseCount++
				}
			}
		}
		metrics["pause_rate"] = MetricResult{Value: float64(pauseCount) / float64(len(intervals)), Calculated: true, SampleSize: len(intervals)}
		metrics["deep_thinking_pause_rate"] = MetricResult{Value: float64(longPauseCount) / float64(len(intervals)), Calculated: true, SampleSize: len(intervals)}
	}

	keyHoldTimes := make([]float64, 0)
	keyDownMap := make(map[string]float64)

	for _, event := range events {
		switch event.Type {
		case "keydown":
			keyDownMap[event.Key] = event.Timestamp
		case "keyup":
			if downTime, exists := keyDownMap[event.Key]; exists {
				holdTime := event.Timestamp - downTime
				if holdTime >= 20 && holdTime <= 1000 {
					keyHoldTimes = append(keyHoldTimes, holdTime)
				}
				delete(keyDownMap, event.Key)
			}
		}
	}

	if len(keyHoldTimes) >= 5 {
		sort.Float64s(keyHoldTimes)
		q1idx := len(keyHoldTimes) / 4
		q3idx := (len(keyHoldTimes) * 3) / 4
		q1 := keyHoldTimes[q1idx]
		q3 := keyHoldTimes[q3idx]
		iqr := q3 - q1

		filteredHoldTimes := make([]float64, 0, len(keyHoldTimes))
		for _, t := range keyHoldTimes {
			if t >= q1-(1.5*iqr) && t <= q3+(1.5*iqr) {
				filteredHoldTimes = append(filteredHoldTimes, t)
			}
		}

		if len(filteredHoldTimes) >= 5 {
			var holdTimeSum float64
			for _, holdTime := range filteredHoldTimes {
				holdTimeSum += holdTime
			}
			avgHoldTime := holdTimeSum / float64(len(filteredHoldTimes))
			metrics["average_key_hold_time"] = MetricResult{Value: avgHoldTime, Calculated: true, SampleSize: len(filteredHoldTimes)}

			var holdTimeVariance float64
			for _, holdTime := range filteredHoldTimes {
				holdTimeVariance += math.Pow(holdTime-avgHoldTime, 2)
			}
			holdTimeVariance /= float64(len(filteredHoldTimes) - 1)
			keyPressVar := math.Sqrt(holdTimeVariance) / avgHoldTime
			metrics["key_press_variability"] = MetricResult{Value: keyPressVar, Calculated: true, SampleSize: len(filteredHoldTimes)}
		}
	}

	if len(keydownEvents) >= 5 {
		correctionCount := 0
		immediateCorrections := 0
		lastCorrection := -1
		charCount := 0

		for i, event := range keydownEvents {
			if event.Key == "Backspace" || event.Key == "Delete" {
				correctionCount++
				if lastCorrection >= 0 && i-lastCorrection <= 3 {
					immediateCorrections++
				}
				lastCorrection = i
			} else if len(event.Key) == 1 || event.Key == "Space" || event.Key == "Enter" {
				charCount++
			}
		}

		if charCount >= 3 {
			metrics["correction_rate"] = MetricResult{Value: float64(correctionCount) / float64(charCount), Calculated: true, SampleSize: charCount}
			if correctionCount > 0 {
				metrics["immediate_correction_tendency"] = MetricResult{Value: float64(immediateCorrections) / float64(correctionCount), Calculated: true, SampleSize: correctionCount}
			}
		}
	}

	if metrics["typing_speed"].Calculated && metrics["average_inter_key_interval"].Calculated && metrics["typing_rhythm_variability"].Calculated {
		typingSpeed := metrics["typing_speed"].Value
		rhythmConsistency := 1.0 / (1.0 + metrics["typing_rhythm_variability"].Value)
		correctionQuality := 1.0

		if metrics["correction_rate"].Calculated {
			correctionQuality = 1.0 / (1.0 + metrics["correction_rate"].Value)
		}

		fluencyScore := 100.0 * ((typingSpeed/5.0)*0.4 + rhythmConsistency*0.4 + correctionQuality*0.2)
		if fluencyScore > 100.0 {
			fluencyScore = 100.0
		}
		metrics["keyboard_fluency"] = MetricResult{Value: fluencyScore, Calculated: true, SampleSize: metrics["typing_speed"].SampleSize}
	}

	return metrics
}

func filterKeyboardEventsByQuestion(questionID *string, interactions *models.InteractionData) []models.KeyboardEvent {
	if questionID == nil {
		return interactions.KeyboardEvents
	}

	filtered := make([]models.KeyboardEvent, 0)
	for _, event := range interactions.KeyboardEvents {
		if event.QuestionID == *questionID {
			filtered = append(filtered, event)
		}
	}

	return filtered
}
