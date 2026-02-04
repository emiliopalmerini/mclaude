package otel

import (
	"os"
	"strconv"
)

// Config holds OTEL exporter configuration.
type Config struct {
	Endpoint string
	Enabled  bool
	Insecure bool
}

// LoadConfig loads OTEL configuration from environment variables.
func LoadConfig() Config {
	enabled, _ := strconv.ParseBool(os.Getenv("MCLAUDE_OTEL_ENABLED"))
	insecure, _ := strconv.ParseBool(os.Getenv("MCLAUDE_OTEL_INSECURE"))

	return Config{
		Endpoint: os.Getenv("MCLAUDE_OTEL_ENDPOINT"),
		Enabled:  enabled,
		Insecure: insecure,
	}
}
