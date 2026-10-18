package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/limber/learning-go-web-notebook/internal/domain"
	"github.com/limber/learning-go-web-notebook/internal/usecase"
)

const maxRequestBodyBytes = 1 << 20 // 1 MiB

// NotebookService 是 HTTP adapter 需要的消费者侧契约，具体实现为
// usecase.NotebookUseCase。
type NotebookService interface {
	Create(ctx context.Context, in usecase.CreateNotebookInput) (*domain.Notebook, error)
	Get(ctx context.Context, id string) (*domain.Notebook, error)
	List(ctx context.Context, in usecase.ListNotebooksInput) ([]*domain.Notebook, int64, error)
	Update(ctx context.Context, id string, in usecase.UpdateNotebookInput) (*domain.Notebook, error)
	Delete(ctx context.Context, id string) error
}

type NotebookHandler struct {
	service  NotebookService
	validate *validator.Validate
	log      *slog.Logger
}

func NewNotebookHandler(service NotebookService, validate *validator.Validate, log *slog.Logger) *NotebookHandler {
	return &NotebookHandler{service: service, validate: validate, log: log}
}

// Create 处理 POST /api/v1/notebooks。
func (h *NotebookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createNotebookRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	notebook, err := h.service.Create(r.Context(), usecase.CreateNotebookInput{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		writeError(w, r, h.log, err)
		return
	}

	w.Header().Set("Location", "/api/v1/notebooks/"+notebook.ID)
	writeJSON(w, http.StatusCreated, dataResponse{Data: toNotebookResponse(notebook)})
}

// List 处理 GET /api/v1/notebooks?limit=&offset=。
func (h *NotebookHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeErrorPayload(w, r, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	limit, offset = usecase.NormalizePagination(limit, offset)

	notebooks, total, err := h.service.List(r.Context(), usecase.ListNotebooksInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, r, h.log, err)
		return
	}

	items := make([]notebookResponse, 0, len(notebooks))
	for _, notebook := range notebooks {
		items = append(items, toNotebookResponse(notebook))
	}

	writeJSON(w, http.StatusOK, listResponse{
		Data: items,
		Meta: listMeta{Limit: limit, Offset: offset, Total: total},
	})
}

// Get 处理 GET /api/v1/notebooks/{id}。
func (h *NotebookHandler) Get(w http.ResponseWriter, r *http.Request) {
	notebook, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, h.log, err)
		return
	}

	writeJSON(w, http.StatusOK, dataResponse{Data: toNotebookResponse(notebook)})
}

// Update 处理 PATCH /api/v1/notebooks/{id}。
func (h *NotebookHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateNotebookRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	if req.Title == nil && req.Content == nil {
		writeErrorPayload(w, r, http.StatusUnprocessableEntity, "validation_failed",
			"request validation failed",
			map[string]string{"body": "at least one of title or content must be provided"},
		)
		return
	}

	notebook, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), usecase.UpdateNotebookInput{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		writeError(w, r, h.log, err)
		return
	}

	writeJSON(w, http.StatusOK, dataResponse{Data: toNotebookResponse(notebook)})
}

// Delete 处理 DELETE /api/v1/notebooks/{id}。
func (h *NotebookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, r, h.log, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// decodeAndValidate 严格解码（拒绝未知字段）、限制请求体大小并执行传输层校验。
func (h *NotebookHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		switch {
		case errors.As(err, &maxBytesErr):
			writeErrorPayload(w, r, http.StatusRequestEntityTooLarge, "request_too_large",
				"request body exceeds 1 MiB", nil)
		case errors.Is(err, io.EOF):
			writeErrorPayload(w, r, http.StatusBadRequest, "invalid_json",
				"request body must not be empty", nil)
		default:
			writeErrorPayload(w, r, http.StatusBadRequest, "invalid_json",
				"invalid JSON: "+err.Error(), nil)
		}
		return false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeErrorPayload(w, r, http.StatusBadRequest, "invalid_json",
			"request body must contain a single JSON object", nil)
		return false
	}

	if err := h.validate.Struct(dst); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			writeErrorPayload(w, r, http.StatusUnprocessableEntity, "validation_failed",
				"request validation failed", validationDetails(validationErrs))
			return false
		}

		writeError(w, r, h.log, err)
		return false
	}

	return true
}

func validationDetails(errs validator.ValidationErrors) map[string]string {
	details := make(map[string]string, len(errs))
	for _, fieldErr := range errs {
		switch fieldErr.Tag() {
		case "required":
			details[fieldErr.Field()] = "is required"
		case "max":
			details[fieldErr.Field()] = fmt.Sprintf("must be at most %s characters", fieldErr.Param())
		default:
			details[fieldErr.Field()] = "failed validation: " + fieldErr.Tag()
		}
	}
	return details
}

func parsePagination(r *http.Request) (int, int, error) {
	limit, err := parseQueryInt(r, "limit")
	if err != nil {
		return 0, 0, err
	}
	offset, err := parseQueryInt(r, "offset")
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

func parseQueryInt(r *http.Request, name string) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return value, nil
}
