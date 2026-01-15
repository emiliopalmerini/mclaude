package app

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr            string        `envconfig:"ADDR" default:":8080"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`

	TursoDatabaseURL string `envconfig:"TURSO_DATABASE_URL_CLAUDE_WATCHER" required:"true"`
	TursoAuthToken   string `envconfig:"TURSO_AUTH_TOKEN_CLAUDE_WATCHER" required:"true"`

	// Pagination
	DefaultPageSize int64 `envconfig:"DEFAULT_PAGE_SIZE" default:"20"`

	// Chart defaults
	DefaultRangeHours int `envconfig:"DEFAULT_RANGE_HOURS" default:"168"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
