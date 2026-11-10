package interfaces

import (
	"context"
	"uuid"

	"github.com/vishenosik/notify/internal/app/common"
)

type APIKeyUseCase interface {
	// Authenticate authenticates a client by API key.
	Authenticate(ctx context.Context, rawKey string) (uuid.UUID, error)
	// Rotate rotates a client's API key.
	Rotate(
		ctx context.Context,
		clientID uuid.UUID,
	) (*common.CreatedClientResult, error)
}
