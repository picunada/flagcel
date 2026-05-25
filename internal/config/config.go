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

	Auth struct {
		OIDCIssuerURL     string        `env:"OIDC_ISSUER_URL" envDefault:""`
		OIDCClientID      string        `env:"OIDC_CLIENT_ID" envDefault:""`
		OIDCClientSecret  string        `env:"OIDC_CLIENT_SECRET" envDefault:""`
		OIDCRedirectURL   string        `env:"OIDC_REDIRECT_URL" envDefault:""`
		AdminEmails       string        `env:"ADMIN_EMAILS" envDefault:""`
		BootstrapEmail    string        `env:"BOOTSTRAP_ADMIN_EMAIL" envDefault:""`
		BootstrapPassword string        `env:"BOOTSTRAP_ADMIN_PASSWORD" envDefault:""`
		BootstrapName     string        `env:"BOOTSTRAP_ADMIN_NAME" envDefault:"Admin"`
		SessionSecret     string        `env:"SESSION_SECRET" envDefault:""`
		CookieSecure      bool          `env:"COOKIE_SECURE" envDefault:"false"`
		SessionTTL        time.Duration `env:"SESSION_TTL" envDefault:"24h"`
	} `envPrefix:"AUTH_"`

	// DebugAddr enables net/http/pprof on this address (e.g. ":6060").
	// Empty disables the debug server. Bind to loopback in any shared env.
	DebugAddr string `env:"DEBUG_ADDR" envDefault:""`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
