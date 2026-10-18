package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config 由环境变量填充。默认值连接本机已有的 ridgewalker_db
// （localhost:5432 上的 postgres/postgres）。
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

	DatabaseURL         string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/ridgewalker_db?sslmode=disable"`
	DatabaseMaxConns    int32  `env:"DATABASE_MAX_CONNS" envDefault:"10"`
	DatabaseAutoMigrate bool   `env:"DATABASE_AUTO_MIGRATE" envDefault:"true"`
}

// Load 在存在 configs/.env 时读取它，然后应用环境变量。缺少 .env 文件没有
// 问题：真实部署直接设置环境变量。
func Load() (Config, error) {
	_ = godotenv.Load("configs/.env")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}
	return cfg, nil
}
