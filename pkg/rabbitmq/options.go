package rabbitmq

import "log/slog"

type Option func(*AmqpClient)

func WithLogger(logger *slog.Logger) Option {
	return func(c *AmqpClient) {
		if logger != nil {
			c.logger = logger
		}
	}
}
