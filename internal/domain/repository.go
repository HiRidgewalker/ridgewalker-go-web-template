package domain

import "context"

// ListNotebooksFilter 描述一次分页请求。
type ListNotebooksFilter struct {
	Limit  int
	Offset int
}

// NotebookRepository 是 notebook 的持久化端口。
//
// 接口放在 domain（由消费者定义契约），由 internal/adapter/postgres 等 adapter
// 实现。use case 只依赖该接口，永远不依赖 pgx 或 SQL。
type NotebookRepository interface {
	Create(ctx context.Context, notebook *Notebook) (*Notebook, error)
	GetByID(ctx context.Context, id string) (*Notebook, error)
	List(ctx context.Context, filter ListNotebooksFilter) ([]*Notebook, int64, error)
	Update(ctx context.Context, notebook *Notebook) (*Notebook, error)
	Delete(ctx context.Context, id string) error
}
