package apikey

import (
	"context"
	"crypto/sha256"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vishenosik/notify/internal/domain/entity"
	repomock "github.com/vishenosik/notify/internal/domain/repository/mock"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func TestService_Authenticate(t *testing.T) {
	t.Parallel()

	clientID := uuid.NewV7()
	plaintext := "ntf_" + strings.Repeat("a", 43)
	hash := HashKey(plaintext)
	saved := entity.NewAPIKey(entity.APIKeyVariants{
		ClientID: clientID,
		Name:     entity.APIKeyDefaultName,
		KeyHash:  hash,
		Prefix:   plaintext[:8],
		Last4:    plaintext[len(plaintext)-4:],
	})

	tests := []struct {
		name    string
		raw     string
		setup   func(keys *repomock.MockAPIKeyRepository)
		want    uuid.UUID
		wantErr error
	}{
		{
			name: "success",
			raw:  "  " + plaintext + "  ",
			setup: func(keys *repomock.MockAPIKeyRepository) {
				keys.EXPECT().GetByHash(mock.Anything, hash).Return(saved, nil).Once()
			},
			want: clientID,
		},
		{
			name:    "missing prefix",
			raw:     "dsk_not-notify",
			wantErr: xerrors.ErrAuth,
		},
		{
			name: "not found is auth",
			raw:  plaintext,
			setup: func(keys *repomock.MockAPIKeyRepository) {
				keys.EXPECT().GetByHash(mock.Anything, hash).Return(nil, xerrors.ErrNotFound).Once()
			},
			wantErr: xerrors.ErrAuth,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			clients := repomock.NewMockClientRepository(t)
			keys := repomock.NewMockAPIKeyRepository(t)
			tx := repomock.NewMockTxManager(t)
			if tc.setup != nil {
				tc.setup(keys)
			}
			got, err := New(clients, keys, tx).Authenticate(context.Background(), tc.raw)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Equal(t, uuid.Nil(), got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestHashKeyIsSHA256OfPlaintext(t *testing.T) {
	t.Parallel()

	raw := "ntf_secret"
	sum := sha256.Sum256([]byte(raw))
	require.Equal(t, sum[:], HashKey(raw))
	require.NotContains(t, string(HashKey(raw)), raw)
}

func TestService_Rotate(t *testing.T) {
	t.Parallel()

	client := entity.NewClient(entity.ClientVariants{Name: "svc"})

	t.Run("archives old key then creates new", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		tx := repomock.NewMockTxManager(t)

		clients.EXPECT().GetByID(mock.Anything, client.ID()).Return(client, nil).Once()
		tx.EXPECT().WithinTx(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}).Once()
		keys.EXPECT().DeleteByClientID(mock.Anything, client.ID(), mock.Anything).
			RunAndReturn(func(_ context.Context, _ uuid.UUID, archivedAt time.Time) error {
				require.False(t, archivedAt.IsZero())
				return nil
			}).Once()
		keys.EXPECT().Create(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, key entity.ValidatedAPIKey) (*entity.APIKey, error) {
				require.Equal(t, client.ID(), key.Value.ClientID())
				require.Equal(t, entity.APIKeyDefaultName, key.Value.Name())
				return key.Value, nil
			}).Once()

		got, err := New(clients, keys, tx).Rotate(context.Background(), client.ID())
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(got.Key, "ntf_"))
		require.Equal(t, client.Name(), got.Name)
		require.Equal(t, HashKey(got.Key), HashKey(got.Key))
	})

	t.Run("client not found", func(t *testing.T) {
		t.Parallel()
		clients := repomock.NewMockClientRepository(t)
		keys := repomock.NewMockAPIKeyRepository(t)
		tx := repomock.NewMockTxManager(t)
		id := uuid.NewV7()
		clients.EXPECT().GetByID(mock.Anything, id).Return(nil, xerrors.ErrNotFound).Once()

		got, err := New(clients, keys, tx).Rotate(context.Background(), id)
		require.ErrorIs(t, err, xerrors.ErrNotFound)
		require.Nil(t, got)
	})

	t.Run("nil id", func(t *testing.T) {
		t.Parallel()
		got, err := New(
			repomock.NewMockClientRepository(t),
			repomock.NewMockAPIKeyRepository(t),
			repomock.NewMockTxManager(t),
		).Rotate(context.Background(), uuid.Nil())
		require.ErrorIs(t, err, xerrors.ErrValidation)
		require.Nil(t, got)
	})
}

func TestGenerateMaterial(t *testing.T) {
	t.Parallel()

	mat, err := GenerateMaterial()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(mat.Plaintext(), keyPrefix))
	require.Len(t, mat.Prefix(), entity.APIKeyPrefixLen)
	require.Len(t, mat.Last4(), entity.APIKeyLast4Len)
	require.Equal(t, HashKey(mat.Plaintext()), mat.Hash())
	require.True(t, strings.HasPrefix(mat.Plaintext(), mat.Prefix()))
	require.True(t, strings.HasSuffix(mat.Plaintext(), mat.Last4()))
}
