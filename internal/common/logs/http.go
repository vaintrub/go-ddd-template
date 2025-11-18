package logs

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	commonctx "github.com/vaintrub/go-ddd-template/internal/common/context"
)

// HTTPLogger creates a middleware that logs HTTP requests with slog
// It logs request start and completion with request ID correlation
func HTTPLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// If X-Request-ID header is present, chi's middleware uses it
			// Otherwise, it generates a UUID
			requestID := middleware.GetReqID(r.Context())

			ctx := commonctx.WithRequestID(r.Context(), requestID)

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			logger.InfoContext(ctx, "START",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			// Process request with updated context
			next.ServeHTTP(ww, r.WithContext(ctx))

			// Calculate duration
			duration := time.Since(start)

			logger.InfoContext(ctx, "END",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status", ww.Status()),
				slog.Duration("duration", duration),
				slog.Int("bytes", ww.BytesWritten()),
			)
		})
	}
}
