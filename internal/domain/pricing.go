package domain

import "time"

type ModelPricing struct {
	ID                          string // e.g., "claude-sonnet-4-20250514"
	DisplayName                 string
	InputPerMillion             float64
	OutputPerMillion            float64
	CacheReadPerMillion         *float64
	CacheWritePerMillion        *float64
	LongContextInputPerMillion  *float64 // Premium pricing for >threshold input tokens
	LongContextOutputPerMillion *float64
	LongContextThreshold        *int64 // Input token threshold (default 200K)
	IsDefault                   bool
	CreatedAt                   time.Time
}

func (p *ModelPricing) CalculateCost(input, output, cacheRead, cacheWrite int64) float64 {
	// Calculate total input tokens (including cache operations) for threshold check
	totalInputTokens := input + cacheRead + cacheWrite

	// Determine if long context pricing applies
	useLongContext := p.LongContextThreshold != nil &&
		p.LongContextInputPerMillion != nil &&
		p.LongContextOutputPerMillion != nil &&
		totalInputTokens > *p.LongContextThreshold

	var inputRate, outputRate float64
	if useLongContext {
		inputRate = *p.LongContextInputPerMillion
		outputRate = *p.LongContextOutputPerMillion
	} else {
		inputRate = p.InputPerMillion
		outputRate = p.OutputPerMillion
	}

	cost := float64(input) * inputRate / 1_000_000
	cost += float64(output) * outputRate / 1_000_000

	if p.CacheReadPerMillion != nil {
		// Cache read pricing is 0.1x base input, scales with long context
		cacheReadRate := *p.CacheReadPerMillion
		if useLongContext {
			// Long context cache read = 0.1 * long context input price
			cacheReadRate = inputRate * 0.1
		}
		cost += float64(cacheRead) * cacheReadRate / 1_000_000
	}
	if p.CacheWritePerMillion != nil {
		// Cache write pricing is 1.25x base input for 5-min cache, scales with long context
		cacheWriteRate := *p.CacheWritePerMillion
		if useLongContext {
			// Long context cache write = 1.25 * long context input price
			cacheWriteRate = inputRate * 1.25
		}
		cost += float64(cacheWrite) * cacheWriteRate / 1_000_000
	}

	return cost
}
