package logs

import (
	"context"
	"log/slog"

	commonctx "github.com/vaintrub/go-ddd-template/internal/common/context"
)

// ContextHandler wraps a slog.Handler and automatically adds
// context attributes (like request ID) to all log records
type ContextHandler struct {
	handler slog.Handler
}

// NewContextHandler creates a handler that extracts attributes from context
func NewContextHandler(handler slog.Handler) *ContextHandler {
	return &ContextHandler{
		handler: handler,
	}
}

// Handle implements slog.Handler
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract request ID from context and add to record
	if requestID := commonctx.RequestIDFromContext(ctx); requestID != "" {
		r.AddAttrs(slog.String("reqid", requestID))
	}

	// Extract user ID from context if present
	if userID := commonctx.UserIDFromContext(ctx); userID != "" {
		r.AddAttrs(slog.String("user_id", userID))
	}

	// Add more context attributes here as needed:
	// - Tenant ID for multi-tenancy
	// - Trace ID for distributed tracing
	// - Session ID for user sessions

	return h.handler.Handle(ctx, r)
}

// Enabled implements slog.Handler
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// WithAttrs implements slog.Handler
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

// WithGroup implements slog.Handler
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{
		handler: h.handler.WithGroup(name),
	}
}
