package domain

// Logger defines the interface for logging
type Logger interface {
	Debug(message string)
	Error(message string)
}
