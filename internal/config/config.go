package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config is populated from environment variables. The defaults match the
// local Docker Compose setup (postgres/postgres on localhost:5433).
type Config struct {
	Env      string `env:"APP_ENV" envDefault:"development"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	HTTPAddr              string        `env:"HTTP_ADDR" envDefault:":8080"`
	HTTPReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	HTTPReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	HTTPWriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"20s"`
	HTTPIdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	HTTPRequestTimeout    time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"15s"`
	ShutdownTimeout       time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`

	DatabaseURL         string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5433/learning_notebook?sslmode=disable"`
	DatabaseMaxConns    int32  `env:"DATABASE_MAX_CONNS" envDefault:"10"`
	DatabaseAutoMigrate bool   `env:"DATABASE_AUTO_MIGRATE" envDefault:"true"`
}

// Load reads configs/.env when present and then applies environment variables.
// A missing .env file is fine: real deployments set variables directly.
func Load() (Config, error) {
	_ = godotenv.Load("configs/.env")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}
	return cfg, nil
}
