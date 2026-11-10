package interfaces

import (
	"context"

	"github.com/vishenosik/notify/internal/app/common"
)

type HealthUseCase interface {
	Check(ctx context.Context) (*common.HealthCheckResponse, error)
}
