package domain

import "time"

type ModelPricing struct {
	ID                   string // e.g., "claude-sonnet-4-20250514"
	DisplayName          string
	InputPerMillion      float64
	OutputPerMillion     float64
	CacheReadPerMillion  *float64
	CacheWritePerMillion *float64
	IsDefault            bool
	CreatedAt            time.Time
}

func (p *ModelPricing) CalculateCost(input, output, cacheRead, cacheWrite int64) float64 {
	cost := float64(input) * p.InputPerMillion / 1_000_000
	cost += float64(output) * p.OutputPerMillion / 1_000_000

	if p.CacheReadPerMillion != nil {
		cost += float64(cacheRead) * *p.CacheReadPerMillion / 1_000_000
	}
	if p.CacheWritePerMillion != nil {
		cost += float64(cacheWrite) * *p.CacheWritePerMillion / 1_000_000
	}

	return cost
}
