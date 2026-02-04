package prometheus

import (
	"os"
	"strconv"
)

// Config holds Prometheus client configuration.
type Config struct {
	URL     string
	Enabled bool
}

// LoadConfig loads Prometheus configuration from environment variables.
func LoadConfig() Config {
	enabled, _ := strconv.ParseBool(os.Getenv("MCLAUDE_PROMETHEUS_ENABLED"))

	return Config{
		URL:     os.Getenv("MCLAUDE_PROMETHEUS_URL"),
		Enabled: enabled,
	}
}
