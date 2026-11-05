package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"uuid"

	"github.com/hashicorp/go-multierror"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
)

var _ repository.Consumer = (*Client)(nil)

// Client is a RabbitMQ adapter for repository.Consumer.
// Reconnect is deferred until this adapter is wired into the application.
type Client struct {
	transport transport
}

type transport interface {
	Consume(ctx context.Context) (<-chan delivery, error)
	Close() error
}

type delivery struct {
	Body []byte
	Ack  func() error
	Nack func(requeue bool) error
}

type pushCommandDTO struct {
	SubscriptionID string            `json:"subscription_id"`
	APIKey         string            `json:"api_key"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	Data           map[string]string `json:"data,omitempty"`
}

// New validates cfg, dials RabbitMQ, and declares topology.
func New(cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("rabbitmq config validation failed: %w", err)
	}

	tr, err := dial(cfg)
	if err != nil {
		return nil, err
	}

	return newClient(tr), nil
}

func newClient(tr transport) *Client {
	return &Client{transport: tr}
}

// Close closes the AMQP channel and then the connection.
func (c *Client) Close() error {
	if err := c.transport.Close(); err != nil {
		return fmt.Errorf("close rabbitmq: %w", err)
	}
	return nil
}

// Consume delivers messages until ctx is cancelled or the broker fails.
// Invalid payloads and repository.ErrDrop are rejected without requeue.
func (c *Client) Consume(ctx context.Context, handler repository.Handler) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	deliveries, err := c.transport.Consume(ctx)
	if err != nil {
		return fmt.Errorf("start consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done(): 
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				if err := ctx.Err(); err != nil {
					return err
				}
				return fmt.Errorf("broker consumer closed")
			}
			if err := handleDelivery(ctx, d, handler); err != nil {
				return err
			}
		}
	}
}

func handleDelivery(ctx context.Context, d delivery, handler repository.Handler) error {
	command, err := unmarshalPushCommand(d.Body)
	if err != nil {
		if nackErr := d.Nack(false); nackErr != nil {
			return fmt.Errorf("nack drop invalid message: %w", nackErr)
		}
		return nil
	}

	err = handler(ctx, command)
	switch {
	case err == nil:
		if ackErr := d.Ack(); ackErr != nil {
			return fmt.Errorf("ack message: %w", ackErr)
		}
	case errors.Is(err, repository.ErrDrop):
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

func unmarshalPushCommand(body []byte) (*entity.PushCommand, error) {
	var dto pushCommandDTO
	if err := json.Unmarshal(body, &dto); err != nil {
		return nil, fmt.Errorf("unmarshal push command: %w", err)
	}

	id, err := uuid.Parse(dto.SubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("parse subscription id: %w", err)
	}

	command := entity.NewPushCommand(id, dto.APIKey, entity.NewPushMessage(dto.Title, dto.Body, dto.Data))
	if err := command.Validate(); err != nil {
		return nil, fmt.Errorf("validate push command: %w", err)
	}
	return command, nil
}

type amqpTransport struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	exchange   string
	queue      string
	routingKey string
}

func dial(cfg *Config) (*amqpTransport, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, multierror.Append(
				fmt.Errorf("open channel: %w", err),
				fmt.Errorf("close connection: %w", closeErr),
			)
		}
		return nil, fmt.Errorf("open channel: %w", err)
	}

	tr := &amqpTransport{
		conn:       conn,
		ch:         ch,
		exchange:   cfg.Exchange,
		queue:      cfg.Queue,
		routingKey: cfg.RoutingKey,
	}
	if err := tr.declare(cfg.Prefetch); err != nil {
		if closeErr := tr.Close(); closeErr != nil {
			return nil, multierror.Append(err, closeErr)
		}
		return nil, err
	}
	return tr, nil
}

func (t *amqpTransport) declare(prefetch int) error {
	if err := t.ch.ExchangeDeclare(t.exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	if _, err := t.ch.QueueDeclare(t.queue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	if err := t.ch.QueueBind(t.queue, t.routingKey, t.exchange, false, nil); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	if err := t.ch.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}
	return nil
}

func (t *amqpTransport) Consume(ctx context.Context) (<-chan delivery, error) {
	msgs, err := t.ch.ConsumeWithContext(ctx, t.queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consume queue: %w", err)
	}

	out := make(chan delivery)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				d := delivery{
					Body: bytes.Clone(msg.Body),
					Ack:  func() error { return msg.Ack(false) },
					Nack: func(requeue bool) error { return msg.Nack(false, requeue) },
				}
				select {
				case <-ctx.Done():
					return
				case out <- d:
				}
			}
		}
	}()
	return out, nil
}

func (t *amqpTransport) Close() error {
	var errs *multierror.Error
	if t.ch != nil {
		if err := t.ch.Close(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("close channel: %w", err))
		}
	}
	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("close connection: %w", err))
		}
	}
	return errs.ErrorOrNil()
}
