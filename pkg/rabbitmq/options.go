package rabbitmq

import (
	"fmt"
	"log/slog"
)

type Option func(*AmqpClient)

func WithLogger(logger *slog.Logger) Option {
	return func(c *AmqpClient) {
		if logger != nil {
			c.logger = logger
		}
	}
}

type stdLogger struct {
	logger *slog.Logger
}

func newStdLogger(logger *slog.Logger) *stdLogger {
	return &stdLogger{
		logger: logger,
	}
}

func (sl *stdLogger) Printf(format string, v ...any) {
	sl.logger.Debug(fmt.Sprintf(format, v...))
}
