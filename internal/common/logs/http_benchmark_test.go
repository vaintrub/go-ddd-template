package logs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	commonctx "github.com/vaintrub/go-ddd-template/internal/common/context"
)

func BenchmarkHTTPMiddlewareOverhead(b *testing.B) {
	logger := Init(config.LoggingConfig{Level: "INFO"})

	// Create router with logging middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(HTTPLogger(logger))
	r.Get("/bench", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Prepare request
	req := httptest.NewRequest("GET", "/bench", nil)
	req.Header.Set("X-Request-ID", "bench-test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// Benchmark without logging middleware (baseline)
func BenchmarkHTTPWithoutMiddleware(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/bench", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/bench", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// Benchmark context operations
func BenchmarkContextOperations(b *testing.B) {
	ctx := commonctx.WithRequestID(context.Background(), "benchmark-id")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = commonctx.RequestIDFromContext(ctx)
	}
}
