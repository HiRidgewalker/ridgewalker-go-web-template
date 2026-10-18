package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limber/learning-go-web-notebook/internal/adapter/postgres/sqlc"
	"github.com/limber/learning-go-web-notebook/internal/domain"
)

// NotebookRepository 用 pgx 与 sqlc 生成的查询实现 domain.NotebookRepository。
type NotebookRepository struct {
	queries *sqlc.Queries
}

var _ domain.NotebookRepository = (*NotebookRepository)(nil)

func NewNotebookRepository(pool *pgxpool.Pool) *NotebookRepository {
	return &NotebookRepository{queries: sqlc.New(pool)}
}

func (r *NotebookRepository) Create(ctx context.Context, notebook *domain.Notebook) (*domain.Notebook, error) {
	row, err := r.queries.CreateNotebook(ctx, sqlc.CreateNotebookParams{
		Title:   notebook.Title,
		Content: notebook.Content,
	})
	if err != nil {
		return nil, mapPostgresError("create notebook", err)
	}
	return toDomainNotebook(row), nil
}

func (r *NotebookRepository) GetByID(ctx context.Context, id string) (*domain.Notebook, error) {
	row, err := r.queries.GetNotebook(ctx, id)
	if err != nil {
		return nil, mapPostgresError("get notebook", err)
	}
	return toDomainNotebook(row), nil
}

func (r *NotebookRepository) List(
	ctx context.Context,
	filter domain.ListNotebooksFilter,
) ([]*domain.Notebook, int64, error) {
	rows, err := r.queries.ListNotebooks(ctx, sqlc.ListNotebooksParams{
		Limit:  int32(filter.Limit),
		Offset: int32(filter.Offset),
	})
	if err != nil {
		return nil, 0, mapPostgresError("list notebooks", err)
	}

	total, err := r.queries.CountNotebooks(ctx)
	if err != nil {
		return nil, 0, mapPostgresError("count notebooks", err)
	}

	notebooks := make([]*domain.Notebook, 0, len(rows))
	for _, row := range rows {
		notebooks = append(notebooks, toDomainNotebook(row))
	}
	return notebooks, total, nil
}

func (r *NotebookRepository) Update(ctx context.Context, notebook *domain.Notebook) (*domain.Notebook, error) {
	row, err := r.queries.UpdateNotebook(ctx, sqlc.UpdateNotebookParams{
		ID:      notebook.ID,
		Title:   notebook.Title,
		Content: notebook.Content,
	})
	if err != nil {
		return nil, mapPostgresError("update notebook", err)
	}
	return toDomainNotebook(row), nil
}

func (r *NotebookRepository) Delete(ctx context.Context, id string) error {
	deleted, err := r.queries.DeleteNotebook(ctx, id)
	if err != nil {
		return mapPostgresError("delete notebook", err)
	}
	if deleted == 0 {
		return fmt.Errorf("delete notebook: %w", domain.ErrNotFound)
	}
	return nil
}

func toDomainNotebook(row sqlc.Notebook) *domain.Notebook {
	return &domain.Notebook{
		ID:        row.ID,
		Title:     row.Title,
		Content:   row.Content,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// mapPostgresError 把驱动层失败转换为 domain 哨兵错误，让上层不必依赖
// PostgreSQL 错误码。
func mapPostgresError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", operation, domain.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "22P02": // invalid_text_representation（例如 UUID 格式错误）
			return fmt.Errorf("%s: %w: %s", operation, domain.ErrValidation, pgErr.Message)
		case "23505": // unique_violation
			return fmt.Errorf("%s: %w", operation, domain.ErrConflict)
		}
	}

	return fmt.Errorf("%s: %w", operation, err)
}
