package entity

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	xerrors "github.com/vishenosik/notify/internal/domain/xerrors"
)

func validPushSubscriptionVariants() PushSubscriptionVariants {
	return PushSubscriptionVariants{
		Endpoint:  "https://push.example/sub",
		P256dh:    "p256dh-key",
		Auth:      "auth-secret",
		UserAgent: "Mozilla/5.0",
	}
}

func TestNewPushSubscriptionSetsFields(t *testing.T) {
	t.Parallel()

	variants := validPushSubscriptionVariants()
	s := NewPushSubscription(variants)

	require.NotEqual(t, uuid.Nil(), s.ID())
	require.False(t, s.CreatedAt().IsZero())
	require.Equal(t, s.CreatedAt(), s.UpdatedAt())
	require.Equal(t, time.UTC, s.CreatedAt().Location())
	require.Equal(t, time.UTC, s.UpdatedAt().Location())
	require.Equal(t, variants.Endpoint, s.Endpoint())
	require.Equal(t, variants.P256dh, s.P256dh())
	require.Equal(t, variants.Auth, s.Auth())
	require.Equal(t, variants.UserAgent, s.UserAgent())
}

func TestCreatedPushSubscriptionRestoresIdentity(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	createdAt := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC)

	s := NewPushSubscription(validPushSubscriptionVariants(), CreatedPushSubscription(Base{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}))

	require.Equal(t, id, s.ID())
	require.Equal(t, createdAt, s.CreatedAt())
	require.Equal(t, updatedAt, s.UpdatedAt())
}

func TestNewValidatedPushSubscription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		variants PushSubscriptionVariants
		wantErr  error
	}{
		{
			name:     "valid",
			variants: validPushSubscriptionVariants(),
		},
		{
			name: "empty user agent allowed",
			variants: PushSubscriptionVariants{
				Endpoint: "https://push.example/sub",
				P256dh:   "p256dh-key",
				Auth:     "auth-secret",
			},
		},
		{
			name: "whitespace user agent allowed",
			variants: PushSubscriptionVariants{
				Endpoint:  "https://push.example/sub",
				P256dh:    "p256dh-key",
				Auth:      "auth-secret",
				UserAgent: "  \t",
			},
		},
		{
			name: "empty endpoint",
			variants: PushSubscriptionVariants{
				P256dh: "p256dh-key",
				Auth:   "auth-secret",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "whitespace endpoint",
			variants: PushSubscriptionVariants{
				Endpoint: "  ",
				P256dh:   "p256dh-key",
				Auth:     "auth-secret",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "empty p256dh",
			variants: PushSubscriptionVariants{
				Endpoint: "https://push.example/sub",
				Auth:     "auth-secret",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "whitespace p256dh",
			variants: PushSubscriptionVariants{
				Endpoint: "https://push.example/sub",
				P256dh:   "\t",
				Auth:     "auth-secret",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "empty auth",
			variants: PushSubscriptionVariants{
				Endpoint: "https://push.example/sub",
				P256dh:   "p256dh-key",
			},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "whitespace auth",
			variants: PushSubscriptionVariants{
				Endpoint: "https://push.example/sub",
				P256dh:   "p256dh-key",
				Auth:     " \n",
			},
			wantErr: xerrors.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewValidated(NewPushSubscription(tc.variants))
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.True(t, got.IsValid())
			require.Equal(t, tc.variants.UserAgent, got.Value.UserAgent())
		})
	}
}

func TestPushSubscriptionIdentityValidation(t *testing.T) {
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
		{
			name:    "zero created at",
			base:    Base{ID: id, UpdatedAt: now},
			wantErr: xerrors.ErrValidation,
		},
		{
			name:    "zero updated at",
			base:    Base{ID: id, CreatedAt: now},
			wantErr: xerrors.ErrValidation,
		},
		{
			name: "created after updated",
			base: Base{
				ID:        id,
				CreatedAt: now.Add(time.Second),
				UpdatedAt: now,
			},
			wantErr: xerrors.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := NewPushSubscription(validPushSubscriptionVariants(), CreatedPushSubscription(tc.base))
			got, err := NewValidated(s)
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
