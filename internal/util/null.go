package util

import "database/sql"

// NullString converts a string to sql.NullString.
// Empty strings are treated as invalid (null).
func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// NullStringPtr converts a *string to sql.NullString.
// Nil pointers are treated as invalid (null).
func NullStringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullStringToPtr converts sql.NullString to *string.
// Invalid values are returned as nil.
func NullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// NullFloat64 converts a *float64 to sql.NullFloat64.
// Nil pointers are treated as invalid (null).
func NullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// NullFloat64Zero converts a *float64 to sql.NullFloat64.
// Nil pointers and zero values are treated as invalid (null).
func NullFloat64Zero(f *float64) sql.NullFloat64 {
	if f == nil || *f == 0 {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// NullInt64 converts a *int64 to sql.NullInt64.
// Nil pointers are treated as invalid (null).
func NullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// BoolToInt64 converts a bool to int64 (true=1, false=0).
// This is useful for SQLite which doesn't have a native boolean type.
func BoolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
