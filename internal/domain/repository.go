package domain

import "context"

// ListNotebooksFilter describes a page request.
type ListNotebooksFilter struct {
	Limit  int
	Offset int
}

// NotebookRepository is the persistence port for notebooks.
//
// The interface lives in the domain (the consumer defines the contract) and is
// implemented by adapters such as internal/adapter/postgres. Use cases depend
// only on this interface, never on pgx or SQL.
type NotebookRepository interface {
	Create(ctx context.Context, notebook *Notebook) (*Notebook, error)
	GetByID(ctx context.Context, id string) (*Notebook, error)
	List(ctx context.Context, filter ListNotebooksFilter) ([]*Notebook, int64, error)
	Update(ctx context.Context, notebook *Notebook) (*Notebook, error)
	Delete(ctx context.Context, id string) error
}
