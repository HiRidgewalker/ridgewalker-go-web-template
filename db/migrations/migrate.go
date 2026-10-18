// Package migrations embeds the goose migration files so the binary can apply
// them without shipping the db/ directory alongside it.
package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

// Run applies all pending up migrations using a short-lived database/sql
// connection. Migrations are wrapped in transactions by goose.
func Run(databaseURL string, log *slog.Logger) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database for migration: %w", err)
	}
	defer func() { _ = db.Close() }()

	db.SetMaxOpenConns(1)

	goose.SetBaseFS(files)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Info("database migrations applied")
	return nil
}
