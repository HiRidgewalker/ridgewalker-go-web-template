package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/limber/learning-go-web-notebook/internal/domain"
	"github.com/limber/learning-go-web-notebook/internal/usecase"
)

// memoryRepository 是单元测试使用的 domain.NotebookRepository 假实现。
type memoryRepository struct {
	items      map[string]*domain.Notebook
	nextID     int
	lastFilter domain.ListNotebooksFilter
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{items: make(map[string]*domain.Notebook)}
}

func (m *memoryRepository) Create(_ context.Context, notebook *domain.Notebook) (*domain.Notebook, error) {
	m.nextID++
	stored := *notebook
	stored.ID = fmt.Sprintf("notebook-%d", m.nextID)
	now := time.Now().UTC()
	stored.CreatedAt = now
	stored.UpdatedAt = now
	m.items[stored.ID] = &stored
	return &stored, nil
}

func (m *memoryRepository) GetByID(_ context.Context, id string) (*domain.Notebook, error) {
	notebook, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("get notebook: %w", domain.ErrNotFound)
	}
	copied := *notebook
	return &copied, nil
}

func (m *memoryRepository) List(_ context.Context, filter domain.ListNotebooksFilter) ([]*domain.Notebook, int64, error) {
	m.lastFilter = filter
	notebooks := make([]*domain.Notebook, 0, len(m.items))
	for _, notebook := range m.items {
		copied := *notebook
		notebooks = append(notebooks, &copied)
	}
	return notebooks, int64(len(notebooks)), nil
}

func (m *memoryRepository) Update(_ context.Context, notebook *domain.Notebook) (*domain.Notebook, error) {
	if _, ok := m.items[notebook.ID]; !ok {
		return nil, fmt.Errorf("update notebook: %w", domain.ErrNotFound)
	}
	stored := *notebook
	m.items[stored.ID] = &stored
	return &stored, nil
}

func (m *memoryRepository) Delete(_ context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return fmt.Errorf("delete notebook: %w", domain.ErrNotFound)
	}
	delete(m.items, id)
	return nil
}

func newUseCase() (*usecase.NotebookUseCase, *memoryRepository) {
	repo := newMemoryRepository()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return usecase.NewNotebookUseCase(repo, log), repo
}

func TestCreateNotebook(t *testing.T) {
	uc, _ := newUseCase()

	created, err := uc.Create(context.Background(), usecase.CreateNotebookInput{
		Title:   "  First note  ",
		Content: "hello",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("Create() returned an empty ID")
	}
	if created.Title != "First note" {
		t.Fatalf("Create() title = %q, want trimmed %q", created.Title, "First note")
	}
}

func TestCreateNotebookRejectsEmptyTitle(t *testing.T) {
	uc, repo := newUseCase()

	_, err := uc.Create(context.Background(), usecase.CreateNotebookInput{Title: "   "})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Create() error = %v, want ErrValidation", err)
	}
	if len(repo.items) != 0 {
		t.Fatalf("Create() persisted %d items for an invalid input", len(repo.items))
	}
}

func TestGetNotebookNotFound(t *testing.T) {
	uc, _ := newUseCase()

	_, err := uc.Get(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestUpdateNotebookPartial(t *testing.T) {
	uc, _ := newUseCase()
	created, err := uc.Create(context.Background(), usecase.CreateNotebookInput{
		Title:   "Original",
		Content: "body",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	newTitle := "Renamed"
	updated, err := uc.Update(context.Background(), created.ID, usecase.UpdateNotebookInput{Title: &newTitle})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "Renamed" {
		t.Fatalf("Update() title = %q, want %q", updated.Title, "Renamed")
	}
	if updated.Content != "body" {
		t.Fatalf("Update() content = %q, want unchanged %q", updated.Content, "body")
	}
}

func TestUpdateNotebookRejectsEmptyTitle(t *testing.T) {
	uc, _ := newUseCase()
	created, err := uc.Create(context.Background(), usecase.CreateNotebookInput{Title: "Original"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	empty := "  "
	_, err = uc.Update(context.Background(), created.ID, usecase.UpdateNotebookInput{Title: &empty})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Update() error = %v, want ErrValidation", err)
	}
}

func TestDeleteNotebookNotFound(t *testing.T) {
	uc, _ := newUseCase()

	err := uc.Delete(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestListNotebooksNormalizesPagination(t *testing.T) {
	uc, repo := newUseCase()

	_, _, err := uc.List(context.Background(), usecase.ListNotebooksInput{Limit: 1000, Offset: -10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.lastFilter.Limit != usecase.MaxPageSize {
		t.Fatalf("List() limit = %d, want %d", repo.lastFilter.Limit, usecase.MaxPageSize)
	}
	if repo.lastFilter.Offset != 0 {
		t.Fatalf("List() offset = %d, want 0", repo.lastFilter.Offset)
	}
}
