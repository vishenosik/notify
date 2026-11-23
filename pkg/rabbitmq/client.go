package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sethvargo/go-retry"
)

type AmqpClient struct {
	conn      *amqp.Connection
	dsn       string
	logger    *slog.Logger
	closeChan chan struct{}
	healthy   atomic.Bool
}

func defaultClient(dsn string) *AmqpClient {
	return &AmqpClient{
		dsn:       dsn,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		closeChan: make(chan struct{}),
	}
}

func NewRabbitMQClient(ctx context.Context, conf *Config, opts ...Option) (*AmqpClient, error) {
	if err := conf.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	ac := defaultClient(conf.DSN)

	for _, opt := range opts {
		opt(ac)
	}
	ac.logger = ac.logger.With(slog.String("component", "rabbitmq"))

	if err := ac.dial(ctx); err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	amqp.SetLogger(newStdLogger(ac.logger))

	return ac, nil
}

func (ac *AmqpClient) Close() error {
	close(ac.closeChan)
	return ac.conn.Close()
}

func (c *AmqpClient) dial(ctx context.Context) error {
	backoff := retry.NewFibonacci(1 * time.Second)
	backoff = retry.WithMaxRetries(10, backoff)

	if err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		conn, err := amqp.DialConfig(c.dsn, amqp.Config{
			Recovery: &amqp.Recovery{},
		})
		if err != nil {
			if isRetryable(err) {
				c.logger.Error("dial rabbitmq error", slog.Any("error", err))
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

	c.logger.Info("rabbitmq connected")
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
		return true
	}
	return false
}

func (ac *AmqpClient) Run(ctx context.Context) error {
	stateChanged := make(chan *amqp.StateChanged, 1)
	ac.conn.NotifyStateChange(stateChanged)

	go ac.statusLoop(stateChanged, ac.logger.With(slog.String("module", "connection")))

	ac.healthy.Store(true)
	return nil
}

func (ac *AmqpClient) statusLoop(ch chan *amqp.StateChanged, logger *slog.Logger) {
	for {
		select {
		case <-ac.closeChan:
			return

		case statusChanged, ok := <-ch:
			if !ok {
				return
			}

			logger.Warn("status changed", slog.Any("status", statusChanged.To.String()))

			switch statusChanged.To {
			case amqp.StateOpen:
				ac.healthy.Store(true)

			case amqp.StateReconnecting:
				logger.Warn("reconnecting to the AMQP server")
				ac.healthy.Store(false)

			case amqp.StateClosed:
				if statusChanged.Err != nil {
					logger.Error("connection closed with error", slog.Any("error", statusChanged.Err))
				} else {
					logger.Info("connection closed normally")
				}
				ac.healthy.Store(false)

			default:
				ac.healthy.Store(false)
			}
		}
	}
}

const (
	exchangeKindDirect = "direct"
)

type Channel struct {
	*amqp.Channel
	closeChan chan struct{}
}

func (ac *AmqpClient) newChannel(conf channelConfig) (*Channel, error) {
	if err := conf.validate(); err != nil {
		return nil, err
	}

	channel, err := ac.conn.Channel()
	if err != nil {
		return nil, err
	}

	chanStateChanged := make(chan *amqp.StateChanged, 1)
	channel.NotifyStateChange(chanStateChanged)

	ch := &Channel{
		Channel:   channel,
		closeChan: make(chan struct{}),
	}

	go func(c chan *amqp.StateChanged) {
		for statusChanged := range c {
			ac.logger.Debug("Status changed",
				slog.String("module", "channel"),
				slog.Any("status", statusChanged),
				slog.String("id", conf.id()),
			)
			switch statusChanged.To {
			case amqp.StateClosed:
				close(ch.closeChan)
			case amqp.StateReconnecting:
				ac.logger.Warn("reconnecting to the AMQP server")
			}
		}
	}(chanStateChanged)

	return ch, nil
}

//
// Pinger implementation
//

func (c *AmqpClient) Name() string {
	return "rabbitmq"
}

func (c *AmqpClient) Ping(ctx context.Context) error {
	if !c.healthy.Load() {
		return fmt.Errorf("rabbitmq is not healthy")
	}
	return nil
}
