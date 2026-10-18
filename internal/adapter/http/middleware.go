package httpadapter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type contextKey int

const requestIDKey contextKey = iota

// RequestIDFromContext returns the request ID stored by the RequestID
// middleware, or an empty string when the middleware is not installed.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// RequestID accepts an incoming X-Request-ID or generates one, then exposes it
// through the request context and response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buf)
}

// AccessLog emits one structured log line per request.
func AccessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				log.InfoContext(r.Context(), "http request",
					"request_id", RequestIDFromContext(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration_ms", time.Since(start).Milliseconds(),
					"remote_addr", r.RemoteAddr,
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// Recoverer converts panics into a 500 response and logs the stack trace.
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				log.ErrorContext(r.Context(), "panic recovered",
					"request_id", RequestIDFromContext(r.Context()),
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
				writeErrorPayload(w, r, http.StatusInternalServerError, "internal_error", "internal server error", nil)
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Timeout bounds the time a handler may spend serving a request.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if d <= 0 {
			return next
		}
		return http.TimeoutHandler(next, d, `{"error":{"code":"timeout","message":"request timed out"}}`)
	}
}
