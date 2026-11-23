package repository

import (
	"context"
	"uuid"

	"github.com/vishenosik/notify/internal/domain/entity"
)

type PushSubscriptionRepository interface {
	// Upsert creates or updates a push subscription.
	Upsert(ctx context.Context, sub entity.ValidatedPushSubscription) (*entity.PushSubscription, error)
	// GetByID retrieves a push subscription by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PushSubscription, error)
	// GetByClientIDAndEndpoint retrieves a push subscription by its client ID and endpoint.
	GetByClientIDAndEndpoint(
		ctx context.Context,
		clientID uuid.UUID,
		endpoint string,
	) (*entity.PushSubscription, error)
	// DeleteByID deletes a push subscription by its ID.
	DeleteByID(ctx context.Context, id uuid.UUID) error
	// DeleteByClientIDAndEndpoint deletes a push subscription by its client ID and endpoint.
	DeleteByClientIDAndEndpoint(
		ctx context.Context,
		clientID uuid.UUID,
		endpoint string,
	) error
	// DeleteByClientID deletes all push subscriptions by its client ID.
	DeleteByClientID(ctx context.Context, clientID uuid.UUID) error
}
