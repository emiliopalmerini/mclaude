package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const htmxKey contextKey = "htmx"

func HTMX(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHTMX := r.Header.Get("HX-Request") == "true"
		ctx := context.WithValue(r.Context(), htmxKey, isHTMX)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsHTMX(r *http.Request) bool {
	if v, ok := r.Context().Value(htmxKey).(bool); ok {
		return v
	}
	return false
}
