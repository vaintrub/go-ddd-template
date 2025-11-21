package decorator

import (
	"context"
	"log/slog"
)

type commandLoggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger *slog.Logger
}

func (d commandLoggingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)

	// Create child logger with command name
	logger := d.logger.With(slog.String("command", handlerType))

	// Log command execution only at DEBUG level
	logger.DebugContext(ctx, "Executing command", slog.Any("command_body", cmd))

	defer func() {
		// Log result only at DEBUG level - HTTP middleware and httperr will handle production logging
		if err == nil {
			logger.DebugContext(ctx, "Command executed successfully")
		} else {
			logger.DebugContext(ctx, "Failed to execute command", slog.Any("error", err))
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type queryLoggingDecorator[C any, R any] struct {
	base   QueryHandler[C, R]
	logger *slog.Logger
}

func (d queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	queryType := generateActionName(cmd)

	// Create child logger with query name
	logger := d.logger.With(slog.String("query", queryType))

	// Log query execution only at DEBUG level
	logger.DebugContext(ctx, "Executing query", slog.Any("query_body", cmd))

	defer func() {
		// Log result only at DEBUG level - HTTP middleware and httperr will handle production logging
		if err == nil {
			logger.DebugContext(ctx, "Query executed successfully")
		} else {
			logger.DebugContext(ctx, "Failed to execute query", slog.Any("error", err))
		}
	}()

	return d.base.Handle(ctx, cmd)
}
