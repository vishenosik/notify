package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherChannel struct {
	*Channel
	logger *slog.Logger
	config *PublisherConfig
}

func (ac *AmqpClient) NewPublisherChannel(conf *PublisherConfig) (*PublisherChannel, error) {
	channel, err := ac.newChannel(conf)
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare(
		conf.Exchange,
		exchangeKindDirect,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	if conf.Queue != "" {
		if _, err := channel.QueueDeclare(
			conf.Queue,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("declare queue: %w", err)
		}
		if err := channel.QueueBind(
			conf.Queue,
			conf.RoutingKey,
			conf.Exchange,
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("bind queue: %w", err)
		}
	}

	return &PublisherChannel{
		Channel: channel,
		logger:  ac.logger.With(slog.String("module", "publisher")),
		config:  conf,
	}, nil
}

// Publish sends a persistent JSON message to the given exchange and routing key.
func (ch *PublisherChannel) Publish(ctx context.Context, body []byte) error {
	if err := ch.PublishWithContext(ctx,
		ch.config.Exchange,
		ch.config.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}
