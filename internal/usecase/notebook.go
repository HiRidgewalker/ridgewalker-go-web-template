package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/limber/learning-go-web-notebook/internal/domain"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// CreateNotebookInput is the use case's own input type. HTTP DTOs are mapped
// to it in the adapter layer, so transport concerns never leak inwards.
type CreateNotebookInput struct {
	Title   string
	Content string
}

// UpdateNotebookInput is a partial update; nil fields are unchanged.
type UpdateNotebookInput struct {
	Title   *string
	Content *string
}

type ListNotebooksInput struct {
	Limit  int
	Offset int
}

// NormalizePagination clamps page parameters to the values the service will
// actually use. It is exported so adapters can report the same values in the
// response metadata.
func NormalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// NotebookUseCase orchestrates notebook operations. It depends only on the
// domain port and the standard library.
type NotebookUseCase struct {
	repo domain.NotebookRepository
	log  *slog.Logger
}

func NewNotebookUseCase(repo domain.NotebookRepository, log *slog.Logger) *NotebookUseCase {
	return &NotebookUseCase{repo: repo, log: log}
}

func (uc *NotebookUseCase) Create(ctx context.Context, in CreateNotebookInput) (*domain.Notebook, error) {
	notebook, err := domain.NewNotebook(in.Title, in.Content)
	if err != nil {
		return nil, err
	}

	created, err := uc.repo.Create(ctx, notebook)
	if err != nil {
		return nil, fmt.Errorf("create notebook: %w", err)
	}

	uc.log.InfoContext(ctx, "notebook created", "notebook_id", created.ID)
	return created, nil
}

func (uc *NotebookUseCase) Get(ctx context.Context, id string) (*domain.Notebook, error) {
	notebook, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get notebook %q: %w", id, err)
	}
	return notebook, nil
}

func (uc *NotebookUseCase) List(ctx context.Context, in ListNotebooksInput) ([]*domain.Notebook, int64, error) {
	in.Limit, in.Offset = NormalizePagination(in.Limit, in.Offset)

	notebooks, total, err := uc.repo.List(ctx, domain.ListNotebooksFilter{
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list notebooks: %w", err)
	}
	return notebooks, total, nil
}

func (uc *NotebookUseCase) Update(ctx context.Context, id string, in UpdateNotebookInput) (*domain.Notebook, error) {
	notebook, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get notebook %q: %w", id, err)
	}

	if err := notebook.Update(in.Title, in.Content); err != nil {
		return nil, err
	}

	updated, err := uc.repo.Update(ctx, notebook)
	if err != nil {
		return nil, fmt.Errorf("update notebook %q: %w", id, err)
	}

	uc.log.InfoContext(ctx, "notebook updated", "notebook_id", updated.ID)
	return updated, nil
}

func (uc *NotebookUseCase) Delete(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete notebook %q: %w", id, err)
	}

	uc.log.InfoContext(ctx, "notebook deleted", "notebook_id", id)
	return nil
}
