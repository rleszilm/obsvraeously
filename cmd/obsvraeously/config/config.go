package config

import (
	"github.com/kelseyhightower/envconfig"
	rest_service "github.com/rleszilm/genms/service/rest"
	"github.com/rleszilm/genms/service/rest/healthcheck"
)

// Auth defines the configuration for the gameday-auth app.
type Auth struct {
	Discord  string              `envconfig:"discord" required:"true"`
	Rest     rest_service.Config `envconfig:"rest"`
	LogLevel string              `envconfig:"log_level" default:"warning"`
	Health   healthcheck.Config
}

// NewFromEnv returns a new GameDay configuration based on environment variables.
func NewFromEnv(prefix string) (*Auth, error) {
	c := Auth{}
	if err := envconfig.Process(prefix, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
