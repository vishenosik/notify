package interfaces

import (
	"context"
	"uuid"

	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
	"github.com/vishenosik/notify/internal/domain/entity"
)

type PushUseCase interface {
	// VAPIDPublicKey returns the VAPID public key for the push service.
	VAPIDPublicKey() (*common.VAPIDPublicKeyResult, error)
	// Subscribe subscribes a client to the push service.
	Subscribe(
		ctx context.Context,
		clientID uuid.UUID,
		cmd *command.SubscribePushCommand,
	) (*common.PushSubscriptionResult, error)
	// Unsubscribes a client from the push service.
	Unsubscribe(
		ctx context.Context,
		clientID uuid.UUID,
		subscriptionID uuid.UUID,
	) error
	// Deliver delivers a push message to a client.
	Deliver(
		ctx context.Context,
		cmd entity.ValidatedPushCommand,
	) error
}
