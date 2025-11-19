package context

import (
	"context"
)

// contextKey is an unexported type for context keys to prevent collisions
type contextKey struct {
	name string
}

// Context keys for various values
var (
	requestIDKey = &contextKey{"request-id"}
	userIDKey    = &contextKey{"user-id"}
	// Add more context keys here as needed:
	// tenantIDKey  = &contextKey{"tenant-id"}
	// traceIDKey   = &contextKey{"trace-id"}
)

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from context
// Returns empty string if not found
func RequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithUserID adds a user ID to the context
// Example of how to extend with additional context values
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts the user ID from context
// Returns empty string if not found
func UserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}
