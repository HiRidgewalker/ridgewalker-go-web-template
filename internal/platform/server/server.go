package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// Server owns the http.Server lifecycle, including graceful shutdown.
type Server struct {
	http            *http.Server
	log             *slog.Logger
	shutdownTimeout time.Duration
}

func New(cfg Config, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		http: &http.Server{
			Addr:              cfg.Addr,
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		log:             log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Run blocks until the server fails or ctx is cancelled, then shuts down
// gracefully. Cancelling ctx is how SIGINT/SIGTERM propagate from main.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server listening", "addr", s.http.Addr)
		errCh <- s.http.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
		s.log.Info("shutting down http server", "timeout", s.shutdownTimeout.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.http.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		s.log.Info("http server stopped")
		return nil
	}
}
