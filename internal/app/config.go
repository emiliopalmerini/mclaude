package app

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr            string        `envconfig:"ADDR" default:":8080"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`

	TursoDatabaseURL string `envconfig:"TURSO_DATABASE_URL" required:"true"`
	TursoAuthToken   string `envconfig:"TURSO_AUTH_TOKEN" required:"true"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
