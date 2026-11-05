package repository

import (
	"context"

	"github.com/vishenosik/notify/internal/domain/entity"
)

// WebPushNotifier delivers a single Web Push notification.
// Dead subscriptions (HTTP 403/404/410) are reported via PushSendResult.Expired;
// the caller is responsible for dropping them.
type WebPushNotifier interface {
	Send(
		ctx context.Context,
		target *entity.PushSubscription,
		message *entity.PushMessage,
	) (*entity.PushSendResult, error)
}
