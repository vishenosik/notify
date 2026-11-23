package client

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/service/apikey"
	"github.com/vishenosik/notify/internal/domain/entity"
	repomock "github.com/vishenosik/notify/internal/domain/repository/mock"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func runTx(t *testing.T, tx *repomock.MockTxManager) {
	t.Helper()
	tx.EXPECT().WithinTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmd     *command.CreateClientCommand
		setup   func(clients *repomock.MockClientRepository, keys *repomock.MockAPIKeyRepository, tx *repomock.MockTxManager)
		wantErr error
	}{
		{
			name: "success stores hash not plaintext",
			cmd: &command.CreateClientCommand{
				Name:        "dashboard-api-gateway",
				Description: "PAU dashboard",
			},
		},
		{
			name:    "nil command",
			wantErr: xerrors.ErrValidation,
		},
		{
			name:    "name required",
			cmd:     &command.CreateClientCommand{},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "conflict",
			cmd:  &command.CreateClientCommand{Name: "dashboard"},
			setup: func(clients *repomock.MockClientRepository, keys *repomock.MockAPIKeyRepository, tx *repomock.MockTxManager) {
				runTx(t, tx)
				clients.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, xerrors.ErrConflict).Once()
			},
			wantErr: xerrors.ErrConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clients := repomock.NewMockClientRepository(t)
			keys := repomock.NewMockAPIKeyRepository(t)
			subs := repomock.NewMockPushSubscriptionRepository(t)
			tx := repomock.NewMockTxManager(t)

			if tc.wantErr == nil {
				runTx(t, tx)
				clients.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, client entity.ValidatedClient) (*entity.Client, error) {
						return client.Value, nil
					}).Once()
				keys.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, key entity.ValidatedAPIKey) (*entity.APIKey, error) {
						require.NotContains(t, string(key.Value.KeyHash()), "ntf_")
						return key.Value, nil
					}).Once()
			} else if tc.setup != nil {
				tc.setup(clients, keys, tx)
			}

			got, err := New(clients, keys, subs, tx).Create(context.Background(), tc.cmd)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.True(t, strings.HasPrefix(got.Key, "ntf_"))
			require.Equal(t, "dashboard-api-gateway", got.Name)
			require.Equal(t, "PAU dashboard", got.Description)
			require.Len(t, got.MaskedAPIKey, len(got.Key[:8])+3+4)
			require.True(t, strings.HasPrefix(got.Key, got.MaskedAPIKey[:8]))
			require.True(t, strings.HasSuffix(got.Key, got.MaskedAPIKey[len(got.MaskedAPIKey)-4:]))
			sum := sha256.Sum256([]byte(got.Key))
			require.NotEqual(t, got.Key, string(sum[:]))
		})
	}
}

func TestService_Create_HashMatchesPlaintext(t *testing.T) {
	t.Parallel()

	var storedHash []byte
	clients := repomock.NewMockClientRepository(t)
	keys := repomock.NewMockAPIKeyRepository(t)
	subs := repomock.NewMockPushSubscriptionRepository(t)
	tx := repomock.NewMockTxManager(t)
	runTx(t, tx)

	clients.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, client entity.ValidatedClient) (*entity.Client, error) {
			return client.Value, nil
		}).Once()
	keys.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, key entity.ValidatedAPIKey) (*entity.APIKey, error) {
			storedHash = key.Value.KeyHash()
			return key.Value, nil
		}).Once()

	got, err := New(clients, keys, subs, tx).Create(context.Background(), &command.CreateClientCommand{Name: "svc"})
	require.NoError(t, err)
	require.Equal(t, apikey.HashKey(got.Key), storedHash)
	require.NotEqual(t, got.Key, string(storedHash))
}

func TestService_Get(t *testing.T) {
	t.Parallel()

	saved := entity.NewClient(entity.ClientVariants{Name: "svc", Description: "d"})
	key := entity.NewAPIKey(entity.APIKeyVariants{
		ClientID: saved.ID(),
		Name:     entity.APIKeyDefaultName,
		KeyHash:  []byte("hash"),
		Prefix:   "ntf_Ab12",
		Last4:    "wxyz",
	})

	t.Run("success with masked key", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		subs := repomock.NewMockPushSubscriptionRepository(t)
		tx := repomock.NewMockTxManager(t)
		clients.EXPECT().GetByID(mock.Anything, saved.ID()).Return(saved, nil).Once()
		keys.EXPECT().GetByClientID(mock.Anything, saved.ID()).Return(key, nil).Once()

		got, err := New(clients, keys, subs, tx).Get(context.Background(), saved.ID())
		require.NoError(t, err)
		require.Equal(t, "ntf_Ab12...wxyz", got.MaskedAPIKey)
		require.Equal(t, saved.Name(), got.Name)
	})

	t.Run("missing key yields empty mask", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		subs := repomock.NewMockPushSubscriptionRepository(t)
		tx := repomock.NewMockTxManager(t)
		clients.EXPECT().GetByID(mock.Anything, saved.ID()).Return(saved, nil).Once()
		keys.EXPECT().GetByClientID(mock.Anything, saved.ID()).Return(nil, xerrors.ErrNotFound).Once()

		got, err := New(clients, keys, subs, tx).Get(context.Background(), saved.ID())
		require.NoError(t, err)
		require.Empty(t, got.MaskedAPIKey)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		subs := repomock.NewMockPushSubscriptionRepository(t)
		tx := repomock.NewMockTxManager(t)
		id := uuid.NewV7()
		clients.EXPECT().GetByID(mock.Anything, id).Return(nil, xerrors.ErrNotFound).Once()

		got, err := New(clients, keys, subs, tx).Get(context.Background(), id)
		require.ErrorIs(t, err, xerrors.ErrNotFound)
		require.Nil(t, got)
	})

	t.Run("nil id", func(t *testing.T) {
		t.Parallel()
		got, err := New(
			repomock.NewMockClientRepository(t),
			repomock.NewMockAPIKeyRepository(t),
			repomock.NewMockPushSubscriptionRepository(t),
			repomock.NewMockTxManager(t),
		).Get(context.Background(), uuid.Nil())
		require.ErrorIs(t, err, xerrors.ErrValidation)
		require.Nil(t, got)
	})
}

func TestService_List(t *testing.T) {
	t.Parallel()

	saved := entity.NewClient(entity.ClientVariants{Name: "svc"})
	clients := repomock.NewMockClientRepository(t)
	keys := repomock.NewMockAPIKeyRepository(t)
	subs := repomock.NewMockPushSubscriptionRepository(t)
	tx := repomock.NewMockTxManager(t)
	clients.EXPECT().List(mock.Anything).Return([]*entity.Client{saved}, nil).Once()
	keys.EXPECT().GetByClientID(mock.Anything, saved.ID()).Return(nil, xerrors.ErrNotFound).Once()

	got, err := New(clients, keys, subs, tx).List(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, saved.Name(), got[0].Name)
	require.Empty(t, got[0].MaskedAPIKey)
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()

	t.Run("archives client keys and subscriptions", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		subs := repomock.NewMockPushSubscriptionRepository(t)
		tx := repomock.NewMockTxManager(t)
		runTx(t, tx)
		clients.EXPECT().Delete(mock.Anything, id, mock.Anything).
			RunAndReturn(func(_ context.Context, _ uuid.UUID, archivedAt time.Time) error {
				require.False(t, archivedAt.IsZero())
				return nil
			}).Once()
		keys.EXPECT().DeleteByClientID(mock.Anything, id, mock.Anything).Return(nil).Once()
		subs.EXPECT().DeleteByClientID(mock.Anything, id).Return(nil).Once()

		require.NoError(t, New(clients, keys, subs, tx).Delete(context.Background(), id))
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		subs := repomock.NewMockPushSubscriptionRepository(t)
		tx := repomock.NewMockTxManager(t)
		runTx(t, tx)
		clients.EXPECT().Delete(mock.Anything, id, mock.Anything).Return(xerrors.ClientNotFoundError()).Once()

		err := New(clients, keys, subs, tx).Delete(context.Background(), id)
		require.ErrorIs(t, err, xerrors.ErrNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		subs := repomock.NewMockPushSubscriptionRepository(t)
		tx := repomock.NewMockTxManager(t)
		runTx(t, tx)
		clients.EXPECT().Delete(mock.Anything, id, mock.Anything).Return(errors.New("db")).Once()

		err := New(clients, keys, subs, tx).Delete(context.Background(), id)
		require.ErrorContains(t, err, "db")
	})
}
