package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/vishenosik/notify/internal/domain/entity"
)

type ClientRepository interface {
	Create(ctx context.Context, client entity.ValidatedClient) (*entity.Client, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Client, error)
	GetByName(ctx context.Context, name string) (*entity.Client, error)
	List(ctx context.Context) ([]*entity.Client, error)
	Delete(ctx context.Context, id uuid.UUID, archivedAt time.Time) error
}
