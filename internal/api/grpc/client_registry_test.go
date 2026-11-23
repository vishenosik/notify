package grpcapi

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
	ifacemock "github.com/vishenosik/notify/internal/app/interface/mock"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func TestClientRegistryCreateClient(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	clients := ifacemock.NewMockClientUseCase(t)
	keys := ifacemock.NewMockAPIKeyUseCase(t)
	clients.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, cmd *command.CreateClientCommand) (*common.CreatedClientResult, error) {
			require.Equal(t, "dashboard", cmd.Name)
			return &common.CreatedClientResult{
				ID:           id,
				Name:         cmd.Name,
				Description:  cmd.Description,
				MaskedAPIKey: "ntf_Ab12...wxyz",
				CreatedAt:    now,
				UpdatedAt:    now,
				Key:          "ntf_plaintext",
			}, nil
		}).Once()

	got, err := newClientRegistryServer(clients, keys).CreateClient(context.Background(), &notifyv1.CreateClientRequest{
		Name:        "dashboard",
		Description: "PAU",
	})
	require.NoError(t, err)
	require.Equal(t, "ntf_plaintext", got.GetKey())
	require.Equal(t, id.String(), got.GetClient().GetId())
	require.Equal(t, "ntf_Ab12...wxyz", got.GetClient().GetMaskedApiKey())
	require.Equal(t, now.Unix(), got.GetClient().GetCreatedAt())
}

func TestClientRegistryGetClientInvalidID(t *testing.T) {
	t.Parallel()

	_, err := newClientRegistryServer(
		ifacemock.NewMockClientUseCase(t),
		ifacemock.NewMockAPIKeyUseCase(t),
	).GetClient(context.Background(), &notifyv1.GetClientRequest{Id: "bad"})
	require.ErrorIs(t, err, xerrors.ErrValidation)
}

func TestClientRegistryListClients(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	clients := ifacemock.NewMockClientUseCase(t)
	keys := ifacemock.NewMockAPIKeyUseCase(t)
	clients.EXPECT().List(mock.Anything).Return([]*common.ClientResult{{
		ID:   id,
		Name: "svc",
	}}, nil).Once()

	got, err := newClientRegistryServer(clients, keys).ListClients(context.Background(), &notifyv1.ListClientsRequest{})
	require.NoError(t, err)
	require.Len(t, got.GetServices(), 1)
	require.Equal(t, id.String(), got.GetServices()[0].GetId())
}

func TestClientRegistryRotateAPIKey(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	clients := ifacemock.NewMockClientUseCase(t)
	keys := ifacemock.NewMockAPIKeyUseCase(t)
	keys.EXPECT().Rotate(mock.Anything, id).Return(&common.CreatedClientResult{
		ID: id, Name: "svc",
		Key: "ntf_rotated",
	}, nil).Once()

	got, err := newClientRegistryServer(clients, keys).RotateAPIKey(context.Background(), &notifyv1.RotateAPIKeyRequest{Id: id.String()})
	require.NoError(t, err)
	require.Equal(t, "ntf_rotated", got.GetKey())
	require.Equal(t, id.String(), got.GetService().GetId())
}
