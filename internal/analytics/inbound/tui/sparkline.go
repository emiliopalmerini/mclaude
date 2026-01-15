package tui

// RenderSparkline creates a simple ASCII sparkline from values using Unicode block characters
func RenderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	// Unicode block characters from lowest to highest
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Find min and max
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Handle edge case where all values are the same
	if max == min {
		result := make([]rune, len(values))
		for i := range result {
			result[i] = blocks[len(blocks)/2]
		}
		return string(result)
	}

	// Map values to block characters
	result := make([]rune, len(values))
	for i, v := range values {
		// Normalize to 0-1 range
		normalized := (v - min) / (max - min)
		// Map to block index (0 to len(blocks)-1)
		idx := int(normalized * float64(len(blocks)-1))
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		result[i] = blocks[idx]
	}

	return string(result)
}
