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

// CreateNotebookInput 是 use case 自己的输入类型。HTTP DTO 在 adapter 层转换为
// 它，传输层关注点不会向内泄漏。
type CreateNotebookInput struct {
	Title   string
	Content string
}

// UpdateNotebookInput 表示部分更新；nil 字段保持不变。
type UpdateNotebookInput struct {
	Title   *string
	Content *string
}

type ListNotebooksInput struct {
	Limit  int
	Offset int
}

// NormalizePagination 把分页参数收敛到服务实际使用的取值。导出它是为了让
// adapter 能在响应元数据中报告同样的值。
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

// NotebookUseCase 编排 notebook 操作，只依赖 domain 端口与标准库。
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

	uc.log.InfoContext(ctx, "已创建笔记", "notebook_id", created.ID)
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

	uc.log.InfoContext(ctx, "已更新笔记", "notebook_id", updated.ID)
	return updated, nil
}

func (uc *NotebookUseCase) Delete(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete notebook %q: %w", id, err)
	}

	uc.log.InfoContext(ctx, "已删除笔记", "notebook_id", id)
	return nil
}
