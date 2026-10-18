package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// New 构造应用 logger。开发环境使用人类可读的文本格式；生产环境使用结构化
// JSON，便于日志采集器解析。
func New(environment, level string) (*slog.Logger, error) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(strings.ToUpper(level))); err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", level, err)
	}

	opts := &slog.HandlerOptions{Level: logLevel}

	var handler slog.Handler
	if environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler).With("service", "learning-go-web-notebook"), nil
}
