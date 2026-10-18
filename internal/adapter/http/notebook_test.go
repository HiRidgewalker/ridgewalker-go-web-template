package httpadapter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpadapter "github.com/limber/learning-go-web-notebook/internal/adapter/http"
	"github.com/limber/learning-go-web-notebook/internal/domain"
	"github.com/limber/learning-go-web-notebook/internal/usecase"
)

type stubNotebookService struct {
	create func(ctx context.Context, in usecase.CreateNotebookInput) (*domain.Notebook, error)
	get    func(ctx context.Context, id string) (*domain.Notebook, error)
	list   func(ctx context.Context, in usecase.ListNotebooksInput) ([]*domain.Notebook, int64, error)
	update func(ctx context.Context, id string, in usecase.UpdateNotebookInput) (*domain.Notebook, error)
	delete func(ctx context.Context, id string) error
}

func (s *stubNotebookService) Create(ctx context.Context, in usecase.CreateNotebookInput) (*domain.Notebook, error) {
	return s.create(ctx, in)
}

func (s *stubNotebookService) Get(ctx context.Context, id string) (*domain.Notebook, error) {
	return s.get(ctx, id)
}

func (s *stubNotebookService) List(
	ctx context.Context,
	in usecase.ListNotebooksInput,
) ([]*domain.Notebook, int64, error) {
	return s.list(ctx, in)
}

func (s *stubNotebookService) Update(
	ctx context.Context,
	id string,
	in usecase.UpdateNotebookInput,
) (*domain.Notebook, error) {
	return s.update(ctx, id, in)
}

func (s *stubNotebookService) Delete(ctx context.Context, id string) error {
	return s.delete(ctx, id)
}

func newTestRouter(t *testing.T, service *stubNotebookService) http.Handler {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := httpadapter.NewNotebookHandler(service, httpadapter.NewValidator(), log)
	return httpadapter.NewRouter(handler, log, httpadapter.RouterOptions{RequestTimeout: 2 * time.Second})
}

func sampleNotebook() *domain.Notebook {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	return &domain.Notebook{
		ID:        "8f14e45f-ceea-467a-9c3f-1b2c3d4e5f60",
		Title:     "First note",
		Content:   "hello",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func performRequest(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateNotebook(t *testing.T) {
	service := &stubNotebookService{
		create: func(_ context.Context, in usecase.CreateNotebookInput) (*domain.Notebook, error) {
			notebook := sampleNotebook()
			notebook.Title = in.Title
			notebook.Content = in.Content
			return notebook, nil
		},
	}
	router := newTestRouter(t, service)

	rec := performRequest(t, router, http.MethodPost, "/api/v1/notebooks",
		`{"title":"First note","content":"hello"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body)
	}
	if location := rec.Header().Get("Location"); location != "/api/v1/notebooks/"+sampleNotebook().ID {
		t.Fatalf("Location = %q", location)
	}

	var response struct {
		Data struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.ID != sampleNotebook().ID || response.Data.Title != "First note" {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
}

func TestCreateNotebookValidationError(t *testing.T) {
	service := &stubNotebookService{
		create: func(context.Context, usecase.CreateNotebookInput) (*domain.Notebook, error) {
			t.Fatal("service must not be called for an invalid request")
			return nil, nil
		},
	}
	router := newTestRouter(t, service)

	rec := performRequest(t, router, http.MethodPost, "/api/v1/notebooks", `{"content":"hello"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnprocessableEntity, rec.Body)
	}

	var response struct {
		Error struct {
			Code    string            `json:"code"`
			Details map[string]string `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Code != "validation_failed" {
		t.Fatalf("error code = %q", response.Error.Code)
	}
	if response.Error.Details["title"] == "" {
		t.Fatalf("missing title detail: %+v", response.Error.Details)
	}
}

func TestCreateNotebookInvalidJSON(t *testing.T) {
	service := &stubNotebookService{}
	router := newTestRouter(t, service)

	rec := performRequest(t, router, http.MethodPost, "/api/v1/notebooks", `{"title":`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body)
	}
}

func TestGetNotebookNotFound(t *testing.T) {
	service := &stubNotebookService{
		get: func(_ context.Context, id string) (*domain.Notebook, error) {
			return nil, fmt.Errorf("get notebook %q: %w", id, domain.ErrNotFound)
		},
	}
	router := newTestRouter(t, service)

	rec := performRequest(t, router, http.MethodGet, "/api/v1/notebooks/missing", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body)
	}

	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want not_found", response.Error.Code)
	}
}

func TestListNotebooksReportsNormalizedPagination(t *testing.T) {
	service := &stubNotebookService{
		list: func(_ context.Context, in usecase.ListNotebooksInput) ([]*domain.Notebook, int64, error) {
			if in.Limit != usecase.MaxPageSize || in.Offset != 0 {
				t.Fatalf("list input = %+v, want limit %d offset 0", in, usecase.MaxPageSize)
			}
			return []*domain.Notebook{sampleNotebook()}, 1, nil
		},
	}
	router := newTestRouter(t, service)

	rec := performRequest(t, router, http.MethodGet, "/api/v1/notebooks?limit=9999&offset=-3", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body)
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Limit  int   `json:"limit"`
			Offset int   `json:"offset"`
			Total  int64 `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.Limit != usecase.MaxPageSize || response.Meta.Offset != 0 || response.Meta.Total != 1 {
		t.Fatalf("unexpected meta: %+v", response.Meta)
	}
	if len(response.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(response.Data))
	}
}

func TestUpdateNotebookRequiresAtLeastOneField(t *testing.T) {
	service := &stubNotebookService{}
	router := newTestRouter(t, service)

	rec := performRequest(t, router, http.MethodPatch, "/api/v1/notebooks/"+sampleNotebook().ID, `{}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnprocessableEntity, rec.Body)
	}
}

func TestDeleteNotebook(t *testing.T) {
	service := &stubNotebookService{
		delete: func(_ context.Context, id string) error {
			if id != sampleNotebook().ID {
				t.Fatalf("delete id = %q", id)
			}
			return nil
		},
	}
	router := newTestRouter(t, service)

	rec := performRequest(t, router, http.MethodDelete, "/api/v1/notebooks/"+sampleNotebook().ID, "")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNoContent, rec.Body)
	}
}

func TestHealthAndReadiness(t *testing.T) {
	service := &stubNotebookService{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := httpadapter.NewNotebookHandler(service, httpadapter.NewValidator(), log)
	router := httpadapter.NewRouter(handler, log, httpadapter.RouterOptions{
		RequestTimeout: time.Second,
		ReadinessCheck: func(context.Context) error { return nil },
	})

	for _, path := range []string{"/healthz", "/readyz"} {
		rec := performRequest(t, router, http.MethodGet, path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusOK)
		}
	}
}
