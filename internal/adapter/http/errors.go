package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/limber/learning-go-web-notebook/internal/domain"
)

type errorEnvelope struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// writeError maps domain sentinel errors to HTTP status codes. Anything that
// is not a known domain error is treated as an internal error and is logged.
func writeError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		writeErrorPayload(w, r, http.StatusUnprocessableEntity, "validation_failed", err.Error(), nil)
	case errors.Is(err, domain.ErrNotFound):
		writeErrorPayload(w, r, http.StatusNotFound, "not_found", "notebook not found", nil)
	case errors.Is(err, domain.ErrConflict):
		writeErrorPayload(w, r, http.StatusConflict, "conflict", "resource conflict", nil)
	default:
		log.ErrorContext(r.Context(), "request failed",
			"request_id", RequestIDFromContext(r.Context()),
			"error", err,
		)
		writeErrorPayload(w, r, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}

func writeErrorPayload(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code, message string,
	details map[string]string,
) {
	writeJSON(w, status, errorEnvelope{Error: errorPayload{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: RequestIDFromContext(r.Context()),
	}})
}
