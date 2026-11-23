package interfaces

import (
	"context"
	"uuid"

	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
)

type ClientUseCase interface {
	// Create creates a new client.
	Create(
		ctx context.Context,
		cmd *command.CreateClientCommand,
	) (*common.CreatedClientResult, error)
	// Get returns a client by ID.
	Get(
		ctx context.Context,
		id uuid.UUID,
	) (*common.ClientResult, error)
	// List returns a list of clients.
	List(ctx context.Context) ([]*common.ClientResult, error)
	// Delete deletes a client by ID.
	Delete(
		ctx context.Context,
		id uuid.UUID,
	) error
}
