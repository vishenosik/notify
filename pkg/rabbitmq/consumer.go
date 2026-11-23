package rabbitmq

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
)

var ErrDropMessage = errors.New("drop message")

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

type ConsumerChannel struct {
	*Channel
	consumer Consumer
	logger   *slog.Logger
}

func (ac *AmqpClient) NewConsumerChannel(
	conf *ConsumerConfig,
	consumer Consumer,
) error {
	channel, err := ac.newChannel(conf)
	if err != nil {
		return err
	}

	if err := declareConsumerChannel(channel, conf); err != nil {
		return err
	}

	ctx := context.Background()

	msgs, err := channel.ConsumeWithContext(ctx, conf.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue: %w", err)
	}
	ch := &ConsumerChannel{
		Channel:  channel,
		consumer: consumer,
	}

	ch.run(ctx, msgs)

	return nil

}

func declareConsumerChannel(channel *Channel, conf *ConsumerConfig) error {
	if err := channel.ExchangeDeclare(
		conf.Exchange,
		exchangeKindDirect,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(
		conf.Queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := channel.QueueBind(
		conf.Queue,
		conf.RoutingKey,
		conf.Exchange,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := channel.Qos(int(conf.Prefetch), 0, false); err != nil {
		return err
	}

	return nil
}

func (c *ConsumerChannel) run(ctx context.Context, msgs <-chan amqp091.Delivery) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case <-c.closeChan:
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

				if err := c.handleDelivery(ctx, d); err != nil {
					c.logger.Error("handle delivery", slog.Any("error", err))
				}
			}
		}
	}()
}

func (c *ConsumerChannel) handleDelivery(ctx context.Context, d delivery) error {
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
