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

// EffectiveRates holds the resolved per-million-token rates after applying
// long-context thresholds and deriving cache rates.
type EffectiveRates struct {
	Input      float64
	Output     float64
	CacheRead  *float64
	CacheWrite *float64
}

// ResolveRates determines the effective per-million-token rates given the
// token counts. Long-context thresholds and derived cache rates (0.1x read,
// 1.25x write) are applied here so callers can store the resolved rates.
func (p *ModelPricing) ResolveRates(input, output, cacheRead, cacheWrite int64) EffectiveRates {
	totalInputTokens := input + cacheRead + cacheWrite

	useLongContext := p.LongContextThreshold != nil &&
		p.LongContextInputPerMillion != nil &&
		p.LongContextOutputPerMillion != nil &&
		totalInputTokens > *p.LongContextThreshold

	var rates EffectiveRates
	if useLongContext {
		rates.Input = *p.LongContextInputPerMillion
		rates.Output = *p.LongContextOutputPerMillion
	} else {
		rates.Input = p.InputPerMillion
		rates.Output = p.OutputPerMillion
	}

	if p.CacheReadPerMillion != nil {
		cr := *p.CacheReadPerMillion
		if useLongContext {
			cr = rates.Input * 0.1
		}
		rates.CacheRead = &cr
	}
	if p.CacheWritePerMillion != nil {
		cw := *p.CacheWritePerMillion
		if useLongContext {
			cw = rates.Input * 1.25
		}
		rates.CacheWrite = &cw
	}

	return rates
}

func (p *ModelPricing) CalculateCost(input, output, cacheRead, cacheWrite int64) float64 {
	rates := p.ResolveRates(input, output, cacheRead, cacheWrite)

	cost := float64(input) * rates.Input / 1_000_000
	cost += float64(output) * rates.Output / 1_000_000
	if rates.CacheRead != nil {
		cost += float64(cacheRead) * *rates.CacheRead / 1_000_000
	}
	if rates.CacheWrite != nil {
		cost += float64(cacheWrite) * *rates.CacheWrite / 1_000_000
	}

	return cost
}
