package logs

import (
	"log/slog"
	"os"
)

// Init initializes the global slog logger with environment-based configuration
func Init() *slog.Logger {
	// Parse log level from environment
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))

	// Configure handler options
	opts := &slog.HandlerOptions{
		AddSource: true, // Include source file locations for debugging
		Level:     level,
	}

	// Create text handler writing to stdout
	baseHandler := slog.NewTextHandler(os.Stdout, opts)

	// Wrap with context handler for automatic request ID injection
	contextHandler := NewContextHandler(baseHandler)

	// Create logger
	logger := slog.New(contextHandler)

	// Set as default for package-level slog calls
	slog.SetDefault(logger)

	return logger
}

// parseLogLevel parses a log level string into slog.Level
// Returns INFO level if parsing fails or input is empty
func parseLogLevel(levelStr string) slog.Level {
	if levelStr == "" {
		return slog.LevelInfo // Default to INFO
	}

	// Use built-in UnmarshalText for robust parsing
	var level slog.Level
	err := level.UnmarshalText([]byte(levelStr))
	if err != nil {
		// Fall back to INFO on parse error
		return slog.LevelInfo
	}

	return level
}
