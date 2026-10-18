package postgres_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"

	migrations "github.com/limber/learning-go-web-notebook/db/migrations"
	"github.com/limber/learning-go-web-notebook/internal/adapter/postgres"
	"github.com/limber/learning-go-web-notebook/internal/domain"
)

// TestNotebookRepositoryIntegration 针对真实 PostgreSQL 实例运行。设置
// TEST_DATABASE_URL 以启用，例如：
//
//	TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/ridgewalker_db?sslmode=disable' \
//	  go test ./internal/adapter/postgres/...
func TestNotebookRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := migrations.Run(dsn, log); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := postgres.NewPool(ctx, dsn, 2)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(ctx, "TRUNCATE notebooks"); err != nil {
		t.Fatalf("truncate notebooks: %v", err)
	}

	repo := postgres.NewNotebookRepository(pool)

	created, err := repo.Create(ctx, &domain.Notebook{Title: "Integration", Content: "hello"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("Create() returned an empty ID")
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != "Integration" || got.Content != "hello" {
		t.Fatalf("GetByID() = %+v", got)
	}

	created.Title = "Updated"
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("Update() title = %q", updated.Title)
	}

	items, total, err := repo.List(ctx, domain.ListNotebooksFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("List() total = %d, items = %d, want 1/1", total, len(items))
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := repo.GetByID(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrNotFound", err)
	}
}
