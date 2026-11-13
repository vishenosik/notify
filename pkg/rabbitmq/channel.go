package rabbitmq

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrDropMessage = errors.New("drop message")

type Channel struct {
	ch         *amqp.Channel
	exchange   string
	queue      string
	routingKey string
	prefetch   uint8
	consumer   Consumer
	logger     *slog.Logger
}

type Delivery struct {
	Body []byte
}

type delivery struct {
	Body []byte
	Ack  func() error
	Nack func(requeue bool) error
}

type Consumer interface {
	Consume(ctx context.Context, delivery Delivery) error
}

func (t *AmqpClient) NewChannel(cfg *ChannelConfig, consumer Consumer) error {
	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("validate channel config: %w", err)
	}

	ch := &Channel{
		exchange:   cfg.Exchange,
		queue:      cfg.Queue,
		routingKey: cfg.RoutingKey,
		prefetch:   cfg.Prefetch,
		consumer:   consumer,
		logger:     slog.Default(),
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.channels[cfg.Queue] = ch

	return nil
}

func (c *Channel) Run(ctx context.Context, conn *amqp.Connection) error {
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	c.ch = channel

	if err := c.declare(); err != nil {
		return fmt.Errorf("declare channel: %w", err)
	}

	msgs, err := c.ch.ConsumeWithContext(ctx, c.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue: %w", err)
	}

	errs := make(chan error)
	defer close(errs)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-msgs:
				if !ok {
					return
				}

				errs := make(chan error)

				go func() {
					defer close(errs)
					d := delivery{
						Body: bytes.Clone(msg.Body),
						Ack:  func() error { return msg.Ack(false) },
						Nack: func(requeue bool) error { return msg.Nack(false, requeue) },
					}

					if err := c.handleDelivery(ctx, d); err != nil {
						errs <- fmt.Errorf("handle delivery: %w", err)
					}
					errs <- nil
				}()

				select {
				case <-ctx.Done():
					return

				case err := <-errs:
					if err != nil {
						c.logger.Error("handle delivery", "error", err)
					}
				}
			}
		}
	}()
	return nil
}

func (c *Channel) declare() error {
	if err := c.ch.ExchangeDeclare(c.exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	if _, err := c.ch.QueueDeclare(c.queue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	if err := c.ch.QueueBind(c.queue, c.routingKey, c.exchange, false, nil); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	if err := c.ch.Qos(int(c.prefetch), 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	return nil
}

func (c *Channel) handleDelivery(ctx context.Context, d delivery) error {
	err := c.consumer.Consume(ctx, Delivery{Body: d.Body})
	switch {
	case err == nil:
		if ackErr := d.Ack(); ackErr != nil {
			return fmt.Errorf("ack message: %w", ackErr)
		}

	case errors.Is(err, ErrDropMessage):
		if nackErr := d.Nack(false); nackErr != nil {
			return fmt.Errorf("nack drop message: %w", nackErr)
		}

	default:
		if nackErr := d.Nack(true); nackErr != nil {
			return fmt.Errorf("nack requeue message: %w", nackErr)
		}
	}
	return nil
}
