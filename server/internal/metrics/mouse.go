package metrics

import (
	"math"
	"sort"

	"crapp-go/internal/models"
)

// calculateClickPrecision calculates average normalized click precision with minimal threshold
func calculateClickPrecision(questionID *string, interactions *models.InteractionData) MetricResult {
	// Filter interactions by question if needed
	inter := filterInteractionsByQuestion(questionID, interactions)

	// Check if we have enough data - need at least 1 interaction
	if len(inter) < 1 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	// Calculate normalized distances
	sum := 0.0
	for _, interaction := range inter {
		// Calculate distance from center
		distX := interaction.ClickX - interaction.TargetX
		distY := interaction.ClickY - interaction.TargetY
		distance := math.Sqrt(distX*distX + distY*distY)

		// Calculate max possible distance (diagonal of target)
		// This is half the diagonal of the element, assuming it's rectangular
		maxDistance := math.Sqrt(math.Pow(interaction.TargetX, 2)+math.Pow(interaction.TargetY, 2)) / 2
		if maxDistance <= 0 {
			maxDistance = 1 // Prevent division by zero
		}

		// Normalized distance (0-1)
		normalizedDistance := distance / maxDistance
		if normalizedDistance > 1 {
			normalizedDistance = 1
		}

		sum += normalizedDistance
	}

	// Calculate precision (higher is better)
	avgNormalizedDistance := sum / float64(len(inter))
	precision := 1 - avgNormalizedDistance

	return MetricResult{
		Value:      precision,
		Calculated: true,
		SampleSize: len(inter),
	}
}

// calculatePathEfficiency calculates mouse path efficiency with improved robustness
func calculatePathEfficiency(questionID *string, interactions *models.InteractionData) MetricResult {
	movements := filterMovementsByQuestion(questionID, interactions)
	inter := filterInteractionsByQuestion(questionID, interactions)

	if len(movements) < 1 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	// Group movements by target/interaction
	targetMovements := make(map[string][]models.MouseMovement)
	for _, movement := range movements {
		if movement.TargetID != "" {
			targetMovements[movement.TargetID] = append(targetMovements[movement.TargetID], movement)
		}
	}

	// Calculate efficiency for each target
	totalEfficiency := 0.0
	count := 0

	for _, interaction := range inter {
		targetID := interaction.TargetID
		relevantMovements := targetMovements[targetID]

		// Need at least 2 movements to calculate a path
		if len(relevantMovements) < 2 {
			continue // Not enough movements to calculate path
		}

		// Sort movements by timestamp
		sort.Slice(relevantMovements, func(i, j int) bool {
			return relevantMovements[i].Timestamp < relevantMovements[j].Timestamp
		})

		// Calculate direct distance (first point to click point)
		firstPoint := relevantMovements[0]
		directDistX := interaction.ClickX - firstPoint.X
		directDistY := interaction.ClickY - firstPoint.Y
		directDistance := math.Sqrt(directDistX*directDistX + directDistY*directDistY)

		// If direct distance is very small, efficiency is meaningless
		if directDistance < 10.0 { // 10 pixels minimum threshold
			continue
		}

		// Calculate actual path distance with filtering for small movements
		actualDistance := 0.0
		lastX, lastY := firstPoint.X, firstPoint.Y

		for i := 1; i < len(relevantMovements); i++ {
			dx := relevantMovements[i].X - lastX
			dy := relevantMovements[i].Y - lastY
			segmentDist := math.Sqrt(dx*dx + dy*dy)

			// Filter out tiny movements that could be noise
			if segmentDist > 1.0 { // 1 pixel minimum threshold
				actualDistance += segmentDist
				lastX, lastY = relevantMovements[i].X, relevantMovements[i].Y
			}
		}

		// Add final segment to click point
		finalDx := interaction.ClickX - lastX
		finalDy := interaction.ClickY - lastY
		finalDist := math.Sqrt(finalDx*finalDx + finalDy*finalDy)

		if finalDist > 1.0 {
			actualDistance += finalDist
		}

		// Calculate efficiency (direct / actual)
		if actualDistance > 0 {
			efficiency := directDistance / actualDistance
			if efficiency > 1.0 {
				efficiency = 1.0 // Cap at 100% efficiency
			}
			totalEfficiency += efficiency
			count++
		}
	}

	if count == 0 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	result := totalEfficiency / float64(count)
	return MetricResult{
		Value:      result,
		Calculated: true,
		SampleSize: count,
	}
}

// calculateOvershootRate calculates the rate of target overshooting with improved sensitivity
func calculateOvershootRate(questionID *string, interactions *models.InteractionData) MetricResult {
	movements := filterMovementsByQuestion(questionID, interactions)
	inter := filterInteractionsByQuestion(questionID, interactions)

	if len(movements) < 1 || len(inter) < 1 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	// Group movements by target
	targetMovements := make(map[string][]models.MouseMovement)
	for _, movement := range movements {
		if movement.TargetID != "" {
			targetMovements[movement.TargetID] = append(targetMovements[movement.TargetID], movement)
		}
	}

	// Calculate overshoot metrics
	overshootSum := 0.0
	totalTargets := 0

	for _, interaction := range inter {
		targetID := interaction.TargetID
		relevantMovements := targetMovements[targetID]

		// Need at least 5 movements for good pattern detection
		if len(relevantMovements) < 5 {
			continue
		}

		// Sort movements by timestamp
		sort.Slice(relevantMovements, func(i, j int) bool {
			return relevantMovements[i].Timestamp < relevantMovements[j].Timestamp
		})

		// Target point (where click occurred)
		targetX := interaction.TargetX
		targetY := interaction.TargetY

		// Track minimum distance to target
		var minDistance float64 = -1 // -1 means not initialized
		var minDistanceIdx int = -1

		// Find the point where we got closest to the target
		for i := 0; i < len(relevantMovements); i++ {
			curr := relevantMovements[i]

			// Distance to target
			dx := curr.X - targetX
			dy := curr.Y - targetY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Track minimum distance to target
			if minDistance < 0 || dist < minDistance {
				minDistance = dist
				minDistanceIdx = i
			}
		}

		// Overshoot score calculation
		overshootScore := 0.0

		// Check if we got close to the target, then moved away (overshoot pattern)
		if minDistanceIdx > 0 && minDistanceIdx < len(relevantMovements)-1 {
			// Calculate how far away we moved after getting closest
			finalDist := math.Sqrt(
				math.Pow(relevantMovements[len(relevantMovements)-1].X-targetX, 2) +
					math.Pow(relevantMovements[len(relevantMovements)-1].Y-targetY, 2))

			// If we ended up further away than our closest approach, likely an overshoot
			if finalDist > minDistance*1.1 { // 10% further away
				// Calculate score based on how much we moved away after getting close
				overshootScore = math.Min(1.0, (finalDist-minDistance)/50.0)
			}
		}

		overshootSum += overshootScore
		totalTargets++
	}

	if totalTargets == 0 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	// Average overshoot score across all targets
	result := overshootSum / float64(totalTargets)
	return MetricResult{
		Value:      result,
		Calculated: true,
		SampleSize: totalTargets,
	}
}

// calculateAverageVelocity calculates average mouse movement velocity with outlier filtering
func calculateAverageVelocity(questionID *string, interactions *models.InteractionData) MetricResult {
	movements := filterMovementsByQuestion(questionID, interactions)

	if len(movements) < 2 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	// Sort movements by timestamp
	sort.Slice(movements, func(i, j int) bool {
		return movements[i].Timestamp < movements[j].Timestamp
	})

	var velocities []float64

	for i := 1; i < len(movements); i++ {
		dx := movements[i].X - movements[i-1].X
		dy := movements[i].Y - movements[i-1].Y
		dt := (movements[i].Timestamp - movements[i-1].Timestamp) / 1000 // Convert to seconds

		if dt > 0 {
			distance := math.Sqrt(dx*dx + dy*dy)

			// Filter out extremely small movements (likely noise)
			if distance < 1.0 {
				continue
			}

			velocity := distance / dt

			// Skip unrealistically high velocities
			if velocity < 10000 {
				velocities = append(velocities, velocity)
			}
		}
	}

	if len(velocities) == 0 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	// Calculate trimmed mean (remove top and bottom 5%) if we have enough samples
	if len(velocities) > 10 {
		sort.Float64s(velocities)
		trimIndex := int(math.Floor(float64(len(velocities)) * 0.05))
		if trimIndex > 0 {
			velocities = velocities[trimIndex : len(velocities)-trimIndex]
		}
	}

	// Calculate mean
	var sum float64
	for _, v := range velocities {
		sum += v
	}
	result := sum / float64(len(velocities))

	return MetricResult{
		Value:      result,
		Calculated: true,
		SampleSize: len(velocities),
	}
}

// calculateVelocityVariability calculates consistency of mouse velocity with outlier handling
func calculateVelocityVariability(questionID *string, interactions *models.InteractionData) MetricResult {
	movements := filterMovementsByQuestion(questionID, interactions)

	if len(movements) < 3 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: 0,
		}
	}

	// Sort movements by timestamp
	sort.Slice(movements, func(i, j int) bool {
		return movements[i].Timestamp < movements[j].Timestamp
	})

	velocities := make([]float64, 0, len(movements)-1)

	for i := 1; i < len(movements); i++ {
		dx := movements[i].X - movements[i-1].X
		dy := movements[i].Y - movements[i-1].Y
		dt := (movements[i].Timestamp - movements[i-1].Timestamp) / 1000 // Convert to seconds

		if dt > 0 {
			distance := math.Sqrt(dx*dx + dy*dy)

			// Filter out very small movements
			if distance < 1.0 {
				continue
			}

			velocity := distance / dt

			// Filter out unrealistic velocities
			if velocity > 0 && velocity < 10000 {
				velocities = append(velocities, velocity)
			}
		}
	}

	if len(velocities) < 3 {
		return MetricResult{
			Value:      0.0,
			Calculated: false,
			SampleSize: len(velocities),
		}
	}

	// Remove outliers using IQR method
	if len(velocities) > 10 {
		sort.Float64s(velocities)
		q1Index := len(velocities) / 4
		q3Index := len(velocities) * 3 / 4
		q1 := velocities[q1Index]
		q3 := velocities[q3Index]
		iqr := q3 - q1
		lowerBound := q1 - 1.5*iqr
		upperBound := q3 + 1.5*iqr

		filteredVelocities := make([]float64, 0, len(velocities))
		for _, v := range velocities {
			if v >= lowerBound && v <= upperBound {
				filteredVelocities = append(filteredVelocities, v)
			}
		}

		// Only use filtered velocities if we didn't filter too many
		if len(filteredVelocities) > len(velocities)/2 {
			velocities = filteredVelocities
		}
	}

	// Calculate average
	var sum float64
	for _, v := range velocities {
		sum += v
	}
	avg := sum / float64(len(velocities))

	// Calculate variance with Bessel's correction
	var variance float64
	for _, v := range velocities {
		variance += math.Pow(v-avg, 2)
	}
	variance /= float64(len(velocities) - 1)

	// Coefficient of variation
	result := math.Sqrt(variance) / avg

	return MetricResult{
		Value:      result,
		Calculated: true,
		SampleSize: len(velocities),
	}
}

func filterMovementsByQuestion(questionID *string, interactions *models.InteractionData) []models.MouseMovement {
	if questionID == nil {
		return interactions.MouseMovements
	}

	filtered := make([]models.MouseMovement, 0)
	for _, movement := range interactions.MouseMovements {
		if movement.QuestionID == *questionID {
			filtered = append(filtered, movement)
		}
	}

	return filtered
}

func filterInteractionsByQuestion(questionID *string, interactions *models.InteractionData) []models.MouseInteraction {
	if questionID == nil {
		// Only return mouse movements (for navigation, etc)
		return interactions.MouseInteractions
	}

	filtered := make([]models.MouseInteraction, 0)
	for _, interaction := range interactions.MouseInteractions {
		if interaction.QuestionID == *questionID {
			filtered = append(filtered, interaction)
		}
	}

	return filtered
}
