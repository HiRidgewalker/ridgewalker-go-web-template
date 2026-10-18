package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// New builds the application logger. Development uses human-readable text;
// production uses structured JSON so log collectors can parse it.
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
