package logs

import (
	"log/slog"
	"os"

	"github.com/vaintrub/go-ddd-template/internal/common/config"
)

// Init initializes the global slog logger with configuration-based settings.
func Init(cfg config.LoggingConfig) *slog.Logger {
	level := parseLogLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     level,
	}

	baseHandler := slog.NewTextHandler(os.Stdout, opts)
	contextHandler := NewContextHandler(baseHandler)

	logger := slog.New(contextHandler)
	slog.SetDefault(logger)

	return logger
}

// parseLogLevel parses a log level string into slog.Level.
// Returns INFO level if parsing fails or input is empty.
func parseLogLevel(levelStr string) slog.Level {
	if levelStr == "" {
		return slog.LevelInfo
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		return slog.LevelInfo
	}

	return level
}
