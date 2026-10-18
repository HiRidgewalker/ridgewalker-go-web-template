package httpadapter

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type RouterOptions struct {
	RequestTimeout time.Duration
	ReadinessCheck func(ctx context.Context) error
}

// NewRouter 装配路由与中间件。中间件顺序很重要：
//
//	RequestID -> AccessLog -> Recoverer -> Timeout -> routes
//
// AccessLog 位于 Recoverer 外层，使恢复后的 panic 能带上最终的 500 状态码
// 被记录。
func NewRouter(handler *NotebookHandler, log *slog.Logger, opts RouterOptions) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestID)
	r.Use(AccessLog(log))
	r.Use(Recoverer(log))
	r.Use(Timeout(opts.RequestTimeout))

	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz(log, opts.ReadinessCheck))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/notebooks", func(r chi.Router) {
			r.Post("/", handler.Create)
			r.Get("/", handler.List)
			r.Get("/{id}", handler.Get)
			r.Patch("/{id}", handler.Update)
			r.Delete("/{id}", handler.Delete)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeErrorPayload(w, r, http.StatusNotFound, "not_found", "route not found", nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeErrorPayload(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	})

	return r
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyz(log *slog.Logger, check func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check == nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := check(ctx); err != nil {
			log.WarnContext(ctx, "就绪检查失败", "error", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
