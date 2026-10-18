package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	migrations "github.com/limber/learning-go-web-notebook/db/migrations"
	httpadapter "github.com/limber/learning-go-web-notebook/internal/adapter/http"
	"github.com/limber/learning-go-web-notebook/internal/adapter/postgres"
	"github.com/limber/learning-go-web-notebook/internal/config"
	"github.com/limber/learning-go-web-notebook/internal/platform/logger"
	"github.com/limber/learning-go-web-notebook/internal/platform/server"
	"github.com/limber/learning-go-web-notebook/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

// run is the composition root: it reads config, wires adapters and use cases
// together, and hands control to the HTTP server. No business logic lives here.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg.Env, cfg.LogLevel)
	if err != nil {
		return err
	}
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.DatabaseAutoMigrate {
		log.Info("applying database migrations")
		if err := migrations.Run(cfg.DatabaseURL, log); err != nil {
			return err
		}
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL, cfg.DatabaseMaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := postgres.NewNotebookRepository(pool)
	notebookUseCase := usecase.NewNotebookUseCase(repo, log)
	notebookHandler := httpadapter.NewNotebookHandler(notebookUseCase, httpadapter.NewValidator(), log)

	router := httpadapter.NewRouter(notebookHandler, log, httpadapter.RouterOptions{
		RequestTimeout: cfg.HTTPRequestTimeout,
		ReadinessCheck: pool.Ping,
	})

	srv := server.New(server.Config{
		Addr:              cfg.HTTPAddr,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		ShutdownTimeout:   cfg.ShutdownTimeout,
	}, router, log)

	return srv.Run(ctx)
}
