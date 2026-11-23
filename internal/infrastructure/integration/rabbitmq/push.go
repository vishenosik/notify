package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"uuid"

	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/domain/entity"
	rabbitmq "github.com/vishenosik/notify/pkg/rabbitmq"
)

const (
	PushSendExchange   = "notify"
	PushSendQueue      = "notify.push.send"
	PushSendRoutingKey = "push.send"
)

var _ rabbitmq.Consumer = (*pushSendBroker)(nil)

type PushSendCommand struct {
	SubscriptionID string            `json:"subscription_id"`
	APIKey         string            `json:"api_key"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	Data           map[string]string `json:"data,omitempty"`
}

type pushSendBroker struct {
	push interfaces.PushUseCase
}

func NewPushSendBroker(
	client *rabbitmq.AmqpClient,
	push interfaces.PushUseCase,
) (*pushSendBroker, error) {
	broker := &pushSendBroker{
		push: push,
	}

	if err := client.NewConsumerChannel(
		&rabbitmq.ConsumerConfig{
			Exchange:   PushSendExchange,
			Queue:      PushSendQueue,
			RoutingKey: PushSendRoutingKey,
			Prefetch:   1,
		},
		broker,
	); err != nil {
		return nil, err
	}

	return broker, nil
}

func (c *pushSendBroker) Consume(ctx context.Context, delivery rabbitmq.Delivery) error {
	var dto PushSendCommand
	if err := json.Unmarshal(delivery.Body, &dto); err != nil {
		return fmt.Errorf("unmarshal push command: %w", err)
	}

	slog.Info("consume push command", slog.Any("dto", dto))

	command, err := entity.NewValidated(
		entity.NewPushCommand(
			uuid.MustParse(dto.SubscriptionID),
			dto.APIKey,
			entity.NewPushMessage(dto.Title, dto.Body, dto.Data),
		),
	)
	if err != nil {
		return rabbitmq.ErrDropMessage
	}

	if err := c.push.Deliver(ctx, command); err != nil {
		if errors.Is(err, rabbitmq.ErrDropMessage) {
			slog.Error("drop push command", slog.Any("error", err))
		}
		return err
	}

	return nil
}
