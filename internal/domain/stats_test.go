package domain

import (
	"math"
	"testing"
)

func TestAggregateStats_ComputeNormalized(t *testing.T) {
	tests := []struct {
		name           string
		stats          AggregateStats
		totalToolCalls int64
		expected       NormalizedMetrics
	}{
		{
			name: "normal case",
			stats: AggregateStats{
				SessionCount:         10,
				TotalTurns:           50,
				TotalTokenInput:      100000,
				TotalTokenOutput:     20000,
				TotalTokenCacheRead:  50000,
				TotalTokenCacheWrite: 10000,
				TotalErrors:          5,
			},
			totalToolCalls: 200,
			expected: NormalizedMetrics{
				TokensPerTurn:    2400,                                     // (100000+20000)/50
				OutputRatio:      0.2,                                      // 20000/100000
				CacheHitRate:     50000.0 / (100000.0 + 50000.0 + 10000.0), // 50000/160000
				ErrorRate:        0.1,                                      // 5/50
				ToolCallsPerTurn: 4.0,                                      // 200/50
			},
		},
		{
			name: "zero turns — all rates zero",
			stats: AggregateStats{
				TotalTurns:       0,
				TotalTokenInput:  1000,
				TotalTokenOutput: 500,
				TotalErrors:      3,
			},
			totalToolCalls: 10,
			expected: NormalizedMetrics{
				TokensPerTurn:    0,
				OutputRatio:      0.5,
				CacheHitRate:     0,
				ErrorRate:        0,
				ToolCallsPerTurn: 0,
			},
		},
		{
			name: "zero input — output ratio zero",
			stats: AggregateStats{
				TotalTurns:       10,
				TotalTokenInput:  0,
				TotalTokenOutput: 500,
				TotalErrors:      0,
			},
			totalToolCalls: 5,
			expected: NormalizedMetrics{
				TokensPerTurn:    50, // (0+500)/10
				OutputRatio:      0,
				CacheHitRate:     0,
				ErrorRate:        0,
				ToolCallsPerTurn: 0.5,
			},
		},
		{
			name: "zero context tokens — cache hit rate zero",
			stats: AggregateStats{
				TotalTurns:           20,
				TotalTokenInput:      0,
				TotalTokenOutput:     1000,
				TotalTokenCacheRead:  0,
				TotalTokenCacheWrite: 0,
				TotalErrors:          0,
			},
			totalToolCalls: 0,
			expected: NormalizedMetrics{
				TokensPerTurn:    50,
				OutputRatio:      0,
				CacheHitRate:     0,
				ErrorRate:        0,
				ToolCallsPerTurn: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.ComputeNormalized(tt.totalToolCalls)
			assertFloatNear(t, "TokensPerTurn", tt.expected.TokensPerTurn, got.TokensPerTurn)
			assertFloatNear(t, "OutputRatio", tt.expected.OutputRatio, got.OutputRatio)
			assertFloatNear(t, "CacheHitRate", tt.expected.CacheHitRate, got.CacheHitRate)
			assertFloatNear(t, "ErrorRate", tt.expected.ErrorRate, got.ErrorRate)
			assertFloatNear(t, "ToolCallsPerTurn", tt.expected.ToolCallsPerTurn, got.ToolCallsPerTurn)
		})
	}
}

func assertFloatNear(t *testing.T, name string, expected, actual float64) {
	t.Helper()
	if math.Abs(expected-actual) > 0.0001 {
		t.Errorf("%s: expected %.6f, got %.6f", name, expected, actual)
	}
}
