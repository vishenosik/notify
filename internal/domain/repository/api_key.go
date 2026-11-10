package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/vishenosik/notify/internal/domain/entity"
)

type APIKeyRepository interface {
	Create(ctx context.Context, key entity.ValidatedAPIKey) (*entity.APIKey, error)
	GetByHash(ctx context.Context, hash []byte) (*entity.APIKey, error)
	GetByClientID(ctx context.Context, clientID uuid.UUID) (*entity.APIKey, error)
	ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*entity.APIKey, error)
	Delete(ctx context.Context, id uuid.UUID, archivedAt time.Time) error
	DeleteByClientID(ctx context.Context, clientID uuid.UUID, archivedAt time.Time) error
}
