package logs

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	commonctx "github.com/vaintrub/go-ddd-template/internal/common/context"
)

func TestContextHandlerInjectsRequestID(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create base handler writing to buffer
	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Wrap with context handler
	contextHandler := NewContextHandler(baseHandler)
	logger := slog.New(contextHandler)

	// Create context with request ID
	ctx := commonctx.WithRequestID(context.Background(), "test-req-456")

	// Log a message
	logger.InfoContext(ctx, "Test message")

	// Verify request ID is in output
	output := buf.String()
	if !strings.Contains(output, "reqid") {
		t.Errorf("Expected output to contain 'reqid', got: %s", output)
	}
	if !strings.Contains(output, "test-req-456") {
		t.Errorf("Expected output to contain 'test-req-456', got: %s", output)
	}
}

// Test: Context without request ID should still log
func TestContextHandlerWithoutRequestID(t *testing.T) {
	var buf bytes.Buffer

	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	contextHandler := NewContextHandler(baseHandler)
	logger := slog.New(contextHandler)

	// Context without request ID
	ctx := context.Background()

	// Log should still work
	logger.InfoContext(ctx, "Test message without reqid")

	output := buf.String()
	if !strings.Contains(output, "Test message without reqid") {
		t.Errorf("Expected message to be logged, got: %s", output)
	}
	// Should not contain reqid field - check for the JSON field pattern
	if strings.Contains(output, `"reqid"`) {
		t.Errorf("Expected output to NOT contain 'reqid' field when not in context, got: %s", output)
	}
}

func TestContextHandlerEnabledDelegates(t *testing.T) {
	// Create handler with INFO level
	baseHandler := slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	contextHandler := NewContextHandler(baseHandler)

	ctx := context.Background()

	// INFO should be enabled
	if !contextHandler.Enabled(ctx, slog.LevelInfo) {
		t.Error("Expected INFO level to be enabled")
	}

	// WARN should be enabled (higher than INFO)
	if !contextHandler.Enabled(ctx, slog.LevelWarn) {
		t.Error("Expected WARN level to be enabled")
	}

	// DEBUG should NOT be enabled (lower than INFO)
	if contextHandler.Enabled(ctx, slog.LevelDebug) {
		t.Error("Expected DEBUG level to be disabled when handler level is INFO")
	}
}

// Test: WithAttrs preserves wrapper
func TestContextHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer

	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	contextHandler := NewContextHandler(baseHandler)

	// Create logger with attributes
	logger := slog.New(contextHandler).With(slog.String("service", "test"))

	ctx := commonctx.WithRequestID(context.Background(), "attr-test-789")

	logger.InfoContext(ctx, "Test with attributes")

	output := buf.String()
	// Should have request ID from context
	if !strings.Contains(output, "attr-test-789") {
		t.Errorf("Expected request ID in output, got: %s", output)
	}
	// Should have service attribute
	if !strings.Contains(output, "service") || !strings.Contains(output, "test") {
		t.Errorf("Expected service attribute in output, got: %s", output)
	}
}

// Test: WithGroup preserves wrapper
func TestContextHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer

	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	contextHandler := NewContextHandler(baseHandler)
	logger := slog.New(contextHandler)

	ctx := commonctx.WithRequestID(context.Background(), "group-test-101")

	// Log with group
	logger.InfoContext(ctx, "Test with group",
		slog.Group("http",
			slog.String("method", "GET"),
			slog.String("path", "/api/test"),
		),
	)

	output := buf.String()
	// Should have request ID
	if !strings.Contains(output, "group-test-101") {
		t.Errorf("Expected request ID in output, got: %s", output)
	}
	// Should have grouped attributes
	if !strings.Contains(output, "http") || !strings.Contains(output, "method") {
		t.Errorf("Expected grouped attributes in output, got: %s", output)
	}
}
