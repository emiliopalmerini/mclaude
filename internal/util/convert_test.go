package util

import (
	"database/sql"
	"testing"
)

func TestToInt64(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int64
	}{
		{"nil", nil, 0},
		{"int64", int64(42), 42},
		{"int", int(7), 7},
		{"float64", float64(3.9), 3},
		{"string valid", "123", 123},
		{"string invalid", "abc", 0},
		{"string empty", "", 0},
		{"NullInt64 valid", sql.NullInt64{Int64: 99, Valid: true}, 99},
		{"NullInt64 null", sql.NullInt64{Valid: false}, 0},
		{"NullFloat64 valid", sql.NullFloat64{Float64: 15.0, Valid: true}, 15},
		{"NullFloat64 null", sql.NullFloat64{Valid: false}, 0},
		{"bool", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToInt64(tt.in); got != tt.want {
				t.Errorf("ToInt64(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want float64
	}{
		{"nil", nil, 0},
		{"float64", float64(3.14), 3.14},
		{"int64", int64(42), 42.0},
		{"int", int(7), 7.0},
		{"string valid", "3.14", 3.14},
		{"string int", "42", 42.0},
		{"string invalid", "abc", 0},
		{"string empty", "", 0},
		{"NullFloat64 valid", sql.NullFloat64{Float64: 3.14, Valid: true}, 3.14},
		{"NullFloat64 null", sql.NullFloat64{Valid: false}, 0},
		{"NullInt64 valid", sql.NullInt64{Int64: 42, Valid: true}, 42.0},
		{"NullInt64 null", sql.NullInt64{Valid: false}, 0},
		{"bool", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToFloat64(tt.in); got != tt.want {
				t.Errorf("ToFloat64(%v) = %f, want %f", tt.in, got, tt.want)
			}
		})
	}
}
