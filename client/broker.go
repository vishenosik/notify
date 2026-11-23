package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"uuid"

	rmq "github.com/vishenosik/notify/internal/infrastructure/integration/rabbitmq"
	"github.com/vishenosik/notify/pkg/rabbitmq"
)

// PushMessage is the browser-visible payload of a single web push.
type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

type pushSendCommand struct {
	SubscriptionID string            `json:"subscription_id"`
	APIKey         string            `json:"api_key"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	Data           map[string]string `json:"data,omitempty"`
}

// BrokerClient is a RabbitMQ publish-only client for push delivery.
type BrokerClient struct {
	apiKey  string
	logger  *slog.Logger
	amqp    *rabbitmq.AmqpClient
	channel *rabbitmq.PublisherChannel
	mu      sync.Mutex
	closed  atomic.Bool
}

type brokerClientOption func(*BrokerClient)

// NewBrokerClient connects to RabbitMQ and declares the push-send topology.
func NewBrokerClient(ctx context.Context, cfg *BrokerConnectionConfig, opts ...brokerClientOption) (*BrokerClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("validate config: config is required")
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	client := &BrokerClient{
		apiKey: cfg.ApiKey,
	}
	for _, opt := range opts {
		opt(client)
	}

	var amqpOpts []rabbitmq.Option
	if client.logger != nil {
		amqpOpts = append(amqpOpts, rabbitmq.WithLogger(client.logger))
	}

	amqpClient, err := rabbitmq.NewRabbitMQClient(ctx, &rabbitmq.Config{DSN: cfg.DSN}, amqpOpts...)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	channel, err := amqpClient.NewPublisherChannel(&rabbitmq.PublisherConfig{
		Exchange:   rmq.PushSendExchange,
		RoutingKey: rmq.PushSendRoutingKey,
		Queue:      rmq.PushSendQueue,
	})
	if err != nil {
		if closeErr := amqpClient.Close(); closeErr != nil {
			return nil, fmt.Errorf("publisher channel: %w (close client: %w)", err, closeErr)
		}
		return nil, fmt.Errorf("publisher channel: %w", err)
	}

	if err := amqpClient.Run(ctx); err != nil {
		if closeErr := amqpClient.Close(); closeErr != nil {
			return nil, fmt.Errorf("run rabbitmq: %w (close client: %w)", err, closeErr)
		}
		return nil, fmt.Errorf("run rabbitmq: %w", err)
	}

	client.amqp = amqpClient
	client.channel = channel
	return client, nil
}

// Send publishes a push-delivery command for the given subscription.
// The API key is taken from the client config, or from WithAPIKey on ctx.
func (c *BrokerClient) Send(ctx context.Context, subscriptionID string, msg PushMessage) error {
	body, err := marshalPushSend(apiKeyFromContext(ctx, c.apiKey), subscriptionID, msg)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Load() {
		return fmt.Errorf("broker client is closed")
	}

	if err := c.channel.Publish(ctx, body); err != nil {
		return fmt.Errorf("publish push command: %w", err)
	}
	return nil
}

func (c *BrokerClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Swap(true) {
		return nil
	}
	return c.amqp.Close()
}

func marshalPushSend(apiKey, subscriptionID string, msg PushMessage) ([]byte, error) {
	subscriptionID = strings.TrimSpace(subscriptionID)
	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription id is required")
	}
	if _, err := uuid.Parse(subscriptionID); err != nil {
		return nil, fmt.Errorf("subscription id: %w", err)
	}

	title := strings.TrimSpace(msg.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	body, err := json.Marshal(pushSendCommand{
		SubscriptionID: subscriptionID,
		APIKey:         apiKey,
		Title:          title,
		Body:           strings.TrimSpace(msg.Body),
		Data:           msg.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal push command: %w", err)
	}
	return body, nil
}
