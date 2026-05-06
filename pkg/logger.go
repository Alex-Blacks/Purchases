package pkg

import (
	"context"
	"log/slog"
)

type Loger struct {
	log *slog.Logger
}

func NewLoger(log *slog.Logger) *Loger {
	return &Loger{log: log}
}

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return logger
}
