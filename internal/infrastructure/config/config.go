package config

import "github.com/kelseyhightower/envconfig"

// Database holds Turso database configuration.
type Database struct {
	URL       string `envconfig:"TURSO_DATABASE_URL" required:"true"`
	AuthToken string `envconfig:"TURSO_AUTH_TOKEN" required:"true"`
}

// Tracker holds configuration for the session tracker.
type Tracker struct {
	Database Database
}

// Dashboard holds configuration for the TUI dashboard.
type Dashboard struct {
	Database        Database
	DefaultPageSize int64 `envconfig:"DEFAULT_PAGE_SIZE" default:"20"`
}

// LoadTracker loads tracker configuration from environment variables.
func LoadTracker() (*Tracker, error) {
	var cfg Tracker
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadDashboard loads dashboard configuration from environment variables.
func LoadDashboard() (*Dashboard, error) {
	var cfg Dashboard
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
