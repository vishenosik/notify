package entity

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	xerrors "github.com/vishenosik/notify/internal/domain/xerrors"
)

func validClientVariants() ClientVariants {
	return ClientVariants{
		Name:        "dashboard-api-gateway",
		Description: "PAU dashboard",
	}
}

func TestNewClientSetsFields(t *testing.T) {
	t.Parallel()

	variants := validClientVariants()
	c := NewClient(variants)

	require.NotEqual(t, uuid.Nil(), c.ID())
	require.False(t, c.CreatedAt().IsZero())
	require.Equal(t, c.CreatedAt(), c.UpdatedAt())
	require.Equal(t, time.UTC, c.CreatedAt().Location())
	require.Equal(t, variants.Name, c.Name())
	require.Equal(t, variants.Description, c.Description())
}

func TestCreatedClientRestoresIdentity(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	createdAt := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC)

	c := NewClient(validClientVariants(), CreatedClient(Base{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}))

	require.Equal(t, id, c.ID())
	require.Equal(t, createdAt, c.CreatedAt())
	require.Equal(t, updatedAt, c.UpdatedAt())
}

func TestNewValidatedClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		variants ClientVariants
		wantErr  error
	}{
		{
			name:     "valid",
			variants: validClientVariants(),
		},
		{
			name: "trims name and description",
			variants: ClientVariants{
				Name:        "  dashboard  ",
				Description: "  desc  ",
			},
		},
		{
			name: "empty description allowed",
			variants: ClientVariants{
				Name: "dashboard",
			},
		},
		{
			name: "empty name",
			variants: ClientVariants{
				Description: "desc",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "whitespace name",
			variants: ClientVariants{
				Name: "  \t",
			},
			wantErr: xerrors.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewValidated(NewClient(tc.variants))
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.True(t, got.IsValid())
			require.Equal(t, got.Value.Name(), got.Value.Name())
		})
	}
}

func TestNewClientTrimsFields(t *testing.T) {
	t.Parallel()

	c := NewClient(ClientVariants{Name: "  dashboard  ", Description: "  desc  "})
	require.Equal(t, "dashboard", c.Name())
	require.Equal(t, "desc", c.Description())
}

func TestClientIdentityValidation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	id := uuid.NewV7()

	tests := []struct {
		name    string
		base    Base
		wantErr error
	}{
		{
			name: "valid",
			base: Base{ID: id, CreatedAt: now, UpdatedAt: now},
		},
		{
			name:    "nil id",
			base:    Base{CreatedAt: now, UpdatedAt: now},
			wantErr: xerrors.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewValidated(NewClient(validClientVariants(), CreatedClient(tc.base)))
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
