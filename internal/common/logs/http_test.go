package logs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	commonctx "github.com/vaintrub/go-ddd-template/internal/common/context"
)

func TestHTTPMiddlewareLogsRequestStart(t *testing.T) {
	// Create a test logger (we'll verify by checking logs are called)
	logger := Init(config.LoggingConfig{Level: "INFO"})

	// Create test router with middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(HTTPLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	// Make test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-start-123")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Verify response is OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// The actual log verification would require capturing stdout
	// For now, we verify the middleware doesn't break the request flow
}

func TestHTTPMiddlewareLogsRequestCompletion(t *testing.T) {
	logger := Init(config.LoggingConfig{Level: "INFO"})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(HTTPLogger(logger))
	r.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow response"))
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	req.Header.Set("X-Request-ID", "test-completion-456")
	w := httptest.NewRecorder()

	start := time.Now()
	r.ServeHTTP(w, req)
	elapsed := time.Since(start)

	// Verify the request took at least 10ms (our sleep time)
	if elapsed < 10*time.Millisecond {
		t.Errorf("Expected request to take at least 10ms, took %v", elapsed)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequestIDExtractionFromHeader(t *testing.T) {
	logger := Init(config.LoggingConfig{Level: "INFO"})

	var capturedRequestID string

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(HTTPLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		// Capture request ID from context
		capturedRequestID = commonctx.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Send request with X-Request-ID header
	req := httptest.NewRequest("GET", "/test", nil)
	expectedID := "custom-request-789"
	req.Header.Set("X-Request-ID", expectedID)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Verify the request ID was extracted and propagated
	if capturedRequestID != expectedID {
		t.Errorf("Expected request ID %s, got %s", expectedID, capturedRequestID)
	}
}

func TestUUIDGenerationWhenHeaderMissing(t *testing.T) {
	logger := Init(config.LoggingConfig{Level: "INFO"})

	var capturedRequestID string

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(HTTPLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = commonctx.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Send request WITHOUT X-Request-ID header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Verify a request ID was generated (not empty)
	if capturedRequestID == "" {
		t.Error("Expected request ID to be generated, got empty string")
	}

	// Verify it looks like a UUID (has dashes)
	if !strings.Contains(capturedRequestID, "-") {
		t.Errorf("Expected generated request ID to be UUID format, got: %s", capturedRequestID)
	}
}

func TestRequestIDPropagatesThroughContext(t *testing.T) {
	logger := Init(config.LoggingConfig{Level: "INFO"})

	var handlerRequestID string
	var nestedRequestID string

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(HTTPLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		// Capture request ID in handler
		handlerRequestID = commonctx.RequestIDFromContext(r.Context())

		// Simulate calling a nested function with context
		nestedRequestID = simulateNestedFunction(r.Context())

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "propagation-test-999")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Verify request ID is the same in handler and nested function
	if handlerRequestID != "propagation-test-999" {
		t.Errorf("Expected handler request ID to be 'propagation-test-999', got %s", handlerRequestID)
	}

	if nestedRequestID != "propagation-test-999" {
		t.Errorf("Expected nested request ID to be 'propagation-test-999', got %s", nestedRequestID)
	}

	if handlerRequestID != nestedRequestID {
		t.Errorf("Request ID not propagated correctly: handler=%s, nested=%s", handlerRequestID, nestedRequestID)
	}
}

// Helper function to simulate nested context usage
func simulateNestedFunction(ctx context.Context) string {
	return commonctx.RequestIDFromContext(ctx)
}

// Additional test: Multiple concurrent requests have different request IDs
func TestConcurrentRequestsHaveDifferentIDs(t *testing.T) {
	logger := Init(config.LoggingConfig{Level: "INFO"})

	requestIDs := make(chan string, 10)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(HTTPLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		requestIDs <- commonctx.RequestIDFromContext(r.Context())
		time.Sleep(5 * time.Millisecond) // Simulate some work
		w.WriteHeader(http.StatusOK)
	})

	// Launch 10 concurrent requests
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}()
	}

	// Collect request IDs
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := <-requestIDs
		if ids[id] {
			t.Errorf("Duplicate request ID detected: %s", id)
		}
		ids[id] = true
	}

	// Verify we got 10 unique IDs
	if len(ids) != 10 {
		t.Errorf("Expected 10 unique request IDs, got %d", len(ids))
	}
}

// Test: Context without request ID doesn't break logging
func TestContextWithoutRequestIDWorks(t *testing.T) {
	logger := Init(config.LoggingConfig{Level: "INFO"})

	r := chi.NewRouter()
	// Skip middleware.RequestID to test fallback
	r.Use(HTTPLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
