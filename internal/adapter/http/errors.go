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

// writeError 把 domain 哨兵错误映射为 HTTP 状态码。不是已知 domain 错误的，
// 按内部错误处理并记录日志。
func writeError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		writeErrorPayload(w, r, http.StatusUnprocessableEntity, "validation_failed", err.Error(), nil)
	case errors.Is(err, domain.ErrNotFound):
		writeErrorPayload(w, r, http.StatusNotFound, "not_found", "notebook not found", nil)
	case errors.Is(err, domain.ErrConflict):
		writeErrorPayload(w, r, http.StatusConflict, "conflict", "resource conflict", nil)
	default:
		log.ErrorContext(r.Context(), "请求处理失败",
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
