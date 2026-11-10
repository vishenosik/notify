package webpush

import (
	"context"

	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
)

type noopSender struct{}

func NewNoopSender() repository.WebPushNotifier {
	return noopSender{}
}

func (noopSender) Send(
	_ context.Context,
	_ *entity.PushSubscription,
	_ *entity.PushMessage,
) (*entity.PushSendResult, error) {
	return &entity.PushSendResult{}, nil
}

var _ repository.WebPushNotifier = noopSender{}
