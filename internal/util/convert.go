package util

import (
	"database/sql"
	"strconv"
)

// ToInt64 safely converts an any to int64.
// Handles int64, int, float64, string, sql.NullInt64, and sql.NullFloat64 types.
// Returns 0 for nil or unsupported types.
func ToInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	case sql.NullInt64:
		if n.Valid {
			return n.Int64
		}
		return 0
	case sql.NullFloat64:
		if n.Valid {
			return int64(n.Float64)
		}
		return 0
	default:
		return 0
	}
}

// ToFloat64 safely converts an any to float64.
// Handles float64, int64, int, string, sql.NullFloat64, and sql.NullInt64 types.
// Returns 0 for nil or unsupported types.
func ToFloat64(v any) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	case sql.NullFloat64:
		if n.Valid {
			return n.Float64
		}
		return 0
	case sql.NullInt64:
		if n.Valid {
			return float64(n.Int64)
		}
		return 0
	default:
		return 0
	}
}
