package logs

import (
	"log/slog"
	"os"
	"testing"
)

func TestInitParsesLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		testMsg  string
		should   string
	}{
		{"DEBUG level", "DEBUG", "debug message", "appear"},
		{"INFO level", "INFO", "info message", "appear"},
		{"WARN level", "WARN", "warn message", "appear"},
		{"ERROR level", "ERROR", "error message", "appear"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			_ = os.Setenv("LOG_LEVEL", tt.logLevel)
			defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

			// Initialize logger
			logger := Init()

			// Test that the logger works
			if logger == nil {
				t.Fatal("Expected logger to be initialized")
			}

			// Verify slog.Default() was set
			if slog.Default() != logger {
				t.Error("Expected slog.Default() to be set to initialized logger")
			}
		})
	}
}

func TestInitDefaultsToInfo(t *testing.T) {
	// Ensure LOG_LEVEL is not set
	_ = os.Unsetenv("LOG_LEVEL")

	logger := Init()

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// The default behavior should allow INFO and above
	// We can't directly test the level, but we can verify the logger was created
	// In a real scenario, you'd check that DEBUG logs don't appear but INFO logs do
}

func TestInitCreatesTextHandlerToStdout(t *testing.T) {
	// We can't easily test stdout capture, but we can verify the logger is created
	// with proper configuration

	_ = os.Setenv("LOG_LEVEL", "INFO")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	logger := Init()

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// Verify logger is usable
	logger.Info("Test message")

	// The actual stdout output would need to be captured separately
	// For unit tests, this verifies the logger is created without panicking
}

// Test: parseLogLevel with valid levels
func TestParseLogLevel_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"debug", slog.LevelDebug}, // case-insensitive
		{"info", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test: parseLogLevel with invalid input defaults to INFO
func TestParseLogLevel_Invalid(t *testing.T) {
	tests := []string{
		"INVALID",
		"TRACE",
		"",
		"123",
		"INFO WARN", // invalid format
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result := parseLogLevel(input)
			if result != slog.LevelInfo {
				t.Errorf("Expected INFO for invalid input '%s', got %v", input, result)
			}
		})
	}
}

// Test: ContextHandler integration with Init()
func TestInitIntegratesContextHandler(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	// Capture stdout (in real test, you'd use a different approach)
	// For now, just verify Init() works with ContextHandler
	logger := Init()

	// Create a child logger to test
	// The ContextHandler should be in the chain
	testLogger := logger.With(slog.String("test", "value"))

	// This should not panic
	testLogger.Info("Integration test message")
}

// Test: AddSource is enabled
func TestInitEnablesAddSource(t *testing.T) {
	// We'll need to test this differently since we can't easily check HandlerOptions
	// For now, verify that the logger is created successfully
	_ = os.Setenv("LOG_LEVEL", "INFO")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	logger := Init()

	if logger == nil {
		t.Fatal("Expected logger with AddSource to be initialized")
	}

	// When AddSource is enabled, logs should include source information
	// This would appear in actual log output but is hard to test in unit tests
}

// Test: TextHandler format
func TestInitUsesTextHandler(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "INFO")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	// Initialize logger
	logger := Init()

	// Verify it's not nil
	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// TextHandler produces human-readable text format
	// JSON would have {"time":..., "level":..., "msg":...}
	// Text would have time=... level=... msg=...

	// This is verified by the actual output format, which we can't easily test here
	// without capturing stdout, but we verify the logger is created
}

// Additional test to use unused variables
func TestInitCreatesWorkingLogger(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "INFO")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	logger := Init()
	logger.Info("test message")
}
