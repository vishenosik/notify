package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sethvargo/go-retry"
)

type AmqpClient struct {
	mu       sync.RWMutex
	conn     *amqp.Connection
	channels map[string]*Channel

	dsn    string
	logger *slog.Logger

	reconnectOnce sync.Once
	closeCh       chan struct{}
}

func NewRabbitMQClient(conf *Config, opts ...Option) (*AmqpClient, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	ac := &AmqpClient{
		dsn:           conf.DSN,
		channels:      make(map[string]*Channel),
		logger:        slog.Default(),
		closeCh:       make(chan struct{}),
		reconnectOnce: sync.Once{},
	}

	for _, opt := range opts {
		opt(ac)
	}

	return ac, nil
}

func (c *AmqpClient) Run(ctx context.Context) error {
	if err := c.dial(ctx); err != nil {
		return err
	}

	for _, ch := range c.channels {
		if err := ch.Run(ctx, c.conn); err != nil {
			return err
		}
	}

	c.reconnectOnce.Do(c.reconnect)

	return nil
}

func (c *AmqpClient) dial(ctx context.Context) error {
	backoff := retry.NewFibonacci(1 * time.Second)
	backoff = retry.WithMaxRetries(10, backoff)

	if err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		conn, err := amqp.Dial(c.dsn)
		if err != nil {
			if isRetryable(err) {
				c.logger.Debug("dial rabbitmq error", slog.Any("error", err))
				return retry.RetryableError(err)
			}
			return fmt.Errorf("dial rabbitmq: %w", err)
		}

		c.conn = conn
		return nil
	}); err != nil {
		c.logger.Error("dial rabbitmq error", slog.Any("error", err))
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	c.logger.Info("dialed rabbitmq")
	return nil
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := errors.AsType[*net.OpError](err); ok {
		return true
	}
	if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	if amqpErr, ok := errors.AsType[*amqp.Error](err); ok {
		slog.Warn("rabbitmq error",
			slog.Any("error", amqpErr.Reason),
			slog.Any("code", amqpErr.Code),
			slog.Any("recover", amqpErr.Recover),
			slog.Any("server_text", amqpErr.Server),
		)
		// return amqpErr.Recover
		return true
	}
	return false
}

func (c *AmqpClient) reconnect() {
	go func() {
		for {
			select {
			case <-c.closeCh:
				c.logger.Info("rabbitmq connection closed")
				return

			case reason, ok := <-c.conn.NotifyClose(make(chan *amqp.Error)):
				if !ok {
					c.logger.Info("rabbitmq connection closed")
					break
				}

				c.logger.Warn("rabbitmq connection closed unexpectedly", slog.Any("reason", reason))

				err := c.Run(context.Background())
				if err != nil {
					c.logger.Error("rabbitmq reconnect failed", slog.Any("error", err))
				}

				c.logger.Info("rabbitmq reconnect success")
			}
		}
	}()
}

func (t *AmqpClient) Close() error {
	close(t.closeCh)

	var errs *multierror.Error
	for _, ch := range t.channels {
		if err := ch.ch.Close(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("close channel: %w", err))
		}
	}

	if err := t.conn.Close(); err != nil {
		if errors.Is(err, amqp.ErrClosed) {
			return nil
		}
		errs = multierror.Append(errs, fmt.Errorf("close connection: %w", err))
	}
	return errs.ErrorOrNil()
}
