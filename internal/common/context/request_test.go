package context

import (
	"context"
	"testing"
)

func TestContextKeyCollisionSafety(t *testing.T) {
	// Create two different context keys
	key1 := &contextKey{"request-id"}
	key2 := &contextKey{"request-id"}

	// They should NOT be equal (pointer comparison)
	if key1 == key2 {
		t.Error("Context keys with same name should not be equal (pointer comparison)")
	}

	// Our actual key should be distinct
	ctx := context.Background()
	ctx = context.WithValue(ctx, key1, "value1")
	ctx = context.WithValue(ctx, requestIDKey, "value2")

	// Values should be stored separately
	if ctx.Value(key1) == ctx.Value(requestIDKey) {
		t.Error("Different context key pointers should store different values")
	}
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-123"

	// Add request ID to context
	ctx = WithRequestID(ctx, requestID)

	// Verify it was stored
	if ctx.Value(requestIDKey) != requestID {
		t.Errorf("Expected request ID %s, got %v", requestID, ctx.Value(requestIDKey))
	}

	// Original context should not be modified
	originalCtx := context.Background()
	_ = WithRequestID(originalCtx, "test")
	if originalCtx.Value(requestIDKey) != nil {
		t.Error("Original context should not be modified")
	}
}

func TestRequestIDFromContext_NotFound(t *testing.T) {
	// Empty context should return empty string
	ctx := context.Background()
	requestID := RequestIDFromContext(ctx)

	if requestID != "" {
		t.Errorf("Expected empty string, got %s", requestID)
	}

	// Context with wrong key should return empty string
	type wrongKeyType string
	ctx = context.WithValue(context.Background(), wrongKeyType("wrong-key"), "some-value")
	requestID = RequestIDFromContext(ctx)

	if requestID != "" {
		t.Errorf("Expected empty string for wrong key, got %s", requestID)
	}

	// Context with wrong type should return empty string
	ctx = context.WithValue(context.Background(), requestIDKey, 123) // int instead of string
	requestID = RequestIDFromContext(ctx)

	if requestID != "" {
		t.Errorf("Expected empty string for wrong type, got %s", requestID)
	}
}

// Additional test: RequestIDFromContext() returns correct value when found
func TestRequestIDFromContext_Found(t *testing.T) {
	ctx := context.Background()
	expectedID := "abc-123-def"
	ctx = WithRequestID(ctx, expectedID)

	actualID := RequestIDFromContext(ctx)

	if actualID != expectedID {
		t.Errorf("Expected %s, got %s", expectedID, actualID)
	}
}
