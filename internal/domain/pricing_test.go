package domain

import (
	"math"
	"testing"
)

func TestModelPricing_CalculateCost_StandardPricing(t *testing.T) {
	pricing := &ModelPricing{
		ID:               "claude-sonnet-4-20250514",
		DisplayName:      "Claude Sonnet 4",
		InputPerMillion:  3.00,
		OutputPerMillion: 15.00,
	}

	// 1000 input, 500 output = $0.003 + $0.0075 = $0.0105
	cost := pricing.CalculateCost(1000, 500, 0, 0)
	expected := 0.0105

	if !floatEquals(cost, expected) {
		t.Errorf("Expected cost %.6f, got %.6f", expected, cost)
	}
}

func TestModelPricing_CalculateCost_WithCache(t *testing.T) {
	cacheRead := 0.30
	cacheWrite := 3.75
	pricing := &ModelPricing{
		ID:                   "claude-sonnet-4-20250514",
		DisplayName:          "Claude Sonnet 4",
		InputPerMillion:      3.00,
		OutputPerMillion:     15.00,
		CacheReadPerMillion:  &cacheRead,
		CacheWritePerMillion: &cacheWrite,
	}

	// 1000 input, 500 output, 100 cache read, 50 cache write
	// $0.003 + $0.0075 + $0.00003 + $0.0001875 = $0.0107175
	cost := pricing.CalculateCost(1000, 500, 100, 50)
	expected := 0.0107175

	if !floatEquals(cost, expected) {
		t.Errorf("Expected cost %.6f, got %.6f", expected, cost)
	}
}

func TestModelPricing_CalculateCost_LongContext_UnderThreshold(t *testing.T) {
	cacheRead := 0.30
	cacheWrite := 3.75
	longInput := 6.00
	longOutput := 22.50
	threshold := int64(200000)

	pricing := &ModelPricing{
		ID:                          "claude-sonnet-4-20250514",
		DisplayName:                 "Claude Sonnet 4",
		InputPerMillion:             3.00,
		OutputPerMillion:            15.00,
		CacheReadPerMillion:         &cacheRead,
		CacheWritePerMillion:        &cacheWrite,
		LongContextInputPerMillion:  &longInput,
		LongContextOutputPerMillion: &longOutput,
		LongContextThreshold:        &threshold,
	}

	// 100K tokens = under threshold, use standard pricing
	// 100000 input, 10000 output = $0.30 + $0.15 = $0.45
	cost := pricing.CalculateCost(100000, 10000, 0, 0)
	expected := 0.45

	if !floatEquals(cost, expected) {
		t.Errorf("Expected cost %.6f, got %.6f", expected, cost)
	}
}

func TestModelPricing_CalculateCost_LongContext_OverThreshold(t *testing.T) {
	cacheRead := 0.30
	cacheWrite := 3.75
	longInput := 6.00
	longOutput := 22.50
	threshold := int64(200000)

	pricing := &ModelPricing{
		ID:                          "claude-sonnet-4-20250514",
		DisplayName:                 "Claude Sonnet 4",
		InputPerMillion:             3.00,
		OutputPerMillion:            15.00,
		CacheReadPerMillion:         &cacheRead,
		CacheWritePerMillion:        &cacheWrite,
		LongContextInputPerMillion:  &longInput,
		LongContextOutputPerMillion: &longOutput,
		LongContextThreshold:        &threshold,
	}

	// 250K tokens = over threshold, use long context pricing
	// 250000 input, 10000 output = $1.50 + $0.225 = $1.725
	cost := pricing.CalculateCost(250000, 10000, 0, 0)
	expected := 1.725

	if !floatEquals(cost, expected) {
		t.Errorf("Expected cost %.6f, got %.6f", expected, cost)
	}
}

func TestModelPricing_CalculateCost_LongContext_CacheCountsTowardThreshold(t *testing.T) {
	cacheRead := 0.30
	cacheWrite := 3.75
	longInput := 6.00
	longOutput := 22.50
	threshold := int64(200000)

	pricing := &ModelPricing{
		ID:                          "claude-sonnet-4-20250514",
		DisplayName:                 "Claude Sonnet 4",
		InputPerMillion:             3.00,
		OutputPerMillion:            15.00,
		CacheReadPerMillion:         &cacheRead,
		CacheWritePerMillion:        &cacheWrite,
		LongContextInputPerMillion:  &longInput,
		LongContextOutputPerMillion: &longOutput,
		LongContextThreshold:        &threshold,
	}

	// 150K input + 30K cache read + 25K cache write = 205K total > 200K threshold
	// Uses long context pricing: $6/MTok input, $22.50/MTok output
	// Cache rates scale: read = 0.1 * 6 = $0.60/MTok, write = 1.25 * 6 = $7.50/MTok
	// 150000 input * $6/MTok = $0.90
	// 10000 output * $22.50/MTok = $0.225
	// 30000 cache read * $0.60/MTok = $0.018
	// 25000 cache write * $7.50/MTok = $0.1875
	// Total = $1.3305
	cost := pricing.CalculateCost(150000, 10000, 30000, 25000)
	expected := 1.3305

	if !floatEquals(cost, expected) {
		t.Errorf("Expected cost %.6f, got %.6f", expected, cost)
	}
}

func TestModelPricing_CalculateCost_Opus46(t *testing.T) {
	cacheRead := 0.50
	cacheWrite := 6.25
	longInput := 10.00
	longOutput := 37.50
	threshold := int64(200000)

	pricing := &ModelPricing{
		ID:                          "claude-opus-4-6-20260115",
		DisplayName:                 "Claude Opus 4.6",
		InputPerMillion:             5.00,
		OutputPerMillion:            25.00,
		CacheReadPerMillion:         &cacheRead,
		CacheWritePerMillion:        &cacheWrite,
		LongContextInputPerMillion:  &longInput,
		LongContextOutputPerMillion: &longOutput,
		LongContextThreshold:        &threshold,
	}

	// Standard pricing: 50K tokens
	// 50000 input * $5/MTok = $0.25
	// 5000 output * $25/MTok = $0.125
	// Total = $0.375
	cost := pricing.CalculateCost(50000, 5000, 0, 0)
	expected := 0.375

	if !floatEquals(cost, expected) {
		t.Errorf("Standard pricing: expected cost %.6f, got %.6f", expected, cost)
	}

	// Long context pricing: 250K tokens
	// 250000 input * $10/MTok = $2.50
	// 5000 output * $37.50/MTok = $0.1875
	// Total = $2.6875
	cost2 := pricing.CalculateCost(250000, 5000, 0, 0)
	expected2 := 2.6875

	if !floatEquals(cost2, expected2) {
		t.Errorf("Long context pricing: expected cost %.6f, got %.6f", expected2, cost2)
	}
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}
