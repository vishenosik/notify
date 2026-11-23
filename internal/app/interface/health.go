package interfaces

import (
	"context"

	"github.com/vishenosik/notify/internal/domain/entity"
)

type HealthUseCase interface {
	Check(ctx context.Context) (*entity.HealthCheck, error)
}
