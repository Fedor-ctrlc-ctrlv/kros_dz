package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"
)

type requestIDKey struct{}

var counter atomic.Uint64

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), counter.Add(1))
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", id)
		rw := &responseWriter{ResponseWriter: w}
		start := time.Now()
		log := logger.With("request_id", id)
		log.Debug("request started", "method", r.Method, "path", r.URL.Path)
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error("unhandled panic", "error", fmt.Sprint(recovered), "stack", string(debug.Stack()))
				if rw.status == 0 {
					http.Error(rw, "internal server error", http.StatusInternalServerError)
				}
			}
			if rw.status == 0 {
				rw.status = http.StatusOK
			}
			log.Info("http request completed", "method", r.Method, "path", r.URL.Path,
				"status", rw.status, "duration_ms", time.Since(start).Milliseconds())
		}()
		next.ServeHTTP(rw, r)
	})
}
