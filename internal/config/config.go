package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port             uint16 `env:"PORT" envDefault:"8080"`
	DatabaseURL      string `env:"DATABASE_URL,required"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat        string `env:"LOG_FORMAT" envDefault:"json"`
	MigrateOnStartup bool   `env:"MIGRATE_ON_STARTUP" envDefault:"true"`

	HTTP struct {
		ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
		WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"5s"`
		IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"10s"`
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
	} `envPrefix:"HTTP_"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
