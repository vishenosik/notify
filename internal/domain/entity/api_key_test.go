package entity

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	xerrors "github.com/vishenosik/notify/internal/domain/xerrors"
)

func validAPIKeyVariants() APIKeyVariants {
	return APIKeyVariants{
		ClientID: uuid.NewV7(),
		Name:     APIKeyDefaultName,
		KeyHash:  []byte("sha256-hash-bytes"),
		Prefix:   "ntf_Ab12",
		Last4:    "wxyz",
	}
}

func TestNewAPIKeySetsFields(t *testing.T) {
	t.Parallel()

	variants := validAPIKeyVariants()
	k := NewAPIKey(variants)

	require.NotEqual(t, uuid.Nil(), k.ID())
	require.Equal(t, variants.ClientID, k.ClientID())
	require.Equal(t, variants.Name, k.Name())
	require.Equal(t, variants.KeyHash, k.KeyHash())
	require.Equal(t, variants.Prefix, k.Prefix())
	require.Equal(t, variants.Last4, k.Last4())
	require.Equal(t, time.UTC, k.CreatedAt().Location())
}

func TestAPIKeyKeyHashIsCloned(t *testing.T) {
	t.Parallel()

	hash := []byte("sha256-hash-bytes")
	k := NewAPIKey(APIKeyVariants{
		ClientID: uuid.NewV7(),
		Name:     APIKeyDefaultName,
		KeyHash:  hash,
		Prefix:   "ntf_Ab12",
		Last4:    "wxyz",
	})
	hash[0] = 'X'
	require.Equal(t, byte('s'), k.KeyHash()[0])

	got := k.KeyHash()
	got[0] = 'Y'
	require.Equal(t, byte('s'), k.KeyHash()[0])
}

func TestCreatedAPIKeyRestoresIdentity(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	createdAt := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC)

	k := NewAPIKey(validAPIKeyVariants(), CreatedAPIKey(Base{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}))

	require.Equal(t, id, k.ID())
	require.Equal(t, createdAt, k.CreatedAt())
	require.Equal(t, updatedAt, k.UpdatedAt())
}

func TestNewValidatedAPIKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		variants APIKeyVariants
		wantErr  error
	}{
		{
			name:     "valid",
			variants: validAPIKeyVariants(),
		},
		{
			name: "name default allowed",
			variants: APIKeyVariants{
				ClientID: uuid.NewV7(),
				Name:     "default",
				KeyHash:  []byte("hash"),
				Prefix:   "ntf_Ab12",
				Last4:    "wxyz",
			},
		},
		{
			name: "nil client id",
			variants: APIKeyVariants{
				Name:    APIKeyDefaultName,
				KeyHash: []byte("hash"),
				Prefix:  "ntf_Ab12",
				Last4:   "wxyz",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "empty name",
			variants: APIKeyVariants{
				ClientID: uuid.NewV7(),
				KeyHash:  []byte("hash"),
				Prefix:   "ntf_Ab12",
				Last4:    "wxyz",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "empty hash",
			variants: APIKeyVariants{
				ClientID: uuid.NewV7(),
				Name:     APIKeyDefaultName,
				Prefix:   "ntf_Ab12",
				Last4:    "wxyz",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "prefix wrong length",
			variants: APIKeyVariants{
				ClientID: uuid.NewV7(),
				Name:     APIKeyDefaultName,
				KeyHash:  []byte("hash"),
				Prefix:   "ntf_",
				Last4:    "wxyz",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "last4 wrong length",
			variants: APIKeyVariants{
				ClientID: uuid.NewV7(),
				Name:     APIKeyDefaultName,
				KeyHash:  []byte("hash"),
				Prefix:   "ntf_Ab12",
				Last4:    "xy",
			},
			wantErr: xerrors.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewValidated(NewAPIKey(tc.variants))
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.True(t, got.IsValid())
		})
	}
}
