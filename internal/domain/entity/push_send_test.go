package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPushTargetValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		target  PushTarget
		wantErr error
	}{
		{
			name:   "valid",
			target: NewPushTarget("https://push.example/sub", "p256dh-key", "auth-secret"),
		},
		{
			name:    "empty endpoint",
			target:  NewPushTarget("", "p256dh-key", "auth-secret"),
			wantErr: ErrPushEndpointRequired,
		},
		{
			name:    "whitespace endpoint",
			target:  NewPushTarget("  ", "p256dh-key", "auth-secret"),
			wantErr: ErrPushEndpointRequired,
		},
		{
			name:    "empty p256dh",
			target:  NewPushTarget("https://push.example/sub", "", "auth-secret"),
			wantErr: ErrPushP256dhRequired,
		},
		{
			name:    "empty auth",
			target:  NewPushTarget("https://push.example/sub", "p256dh-key", ""),
			wantErr: ErrPushAuthRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.target.Validate()
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "https://push.example/sub", tc.target.Endpoint())
			require.Equal(t, "p256dh-key", tc.target.P256dh())
			require.Equal(t, "auth-secret", tc.target.Auth())
		})
	}
}

func TestPushMessageValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid with data", func(t *testing.T) {
		t.Parallel()
		msg := NewPushMessage("Hello", "World", map[string]string{"url": "/inbox"})
		require.NoError(t, msg.Validate())
		require.Equal(t, "Hello", msg.Title())
		require.Equal(t, "World", msg.Body())
		require.Equal(t, map[string]string{"url": "/inbox"}, msg.Data())
	})

	t.Run("empty body allowed", func(t *testing.T) {
		t.Parallel()
		msg := NewPushMessage("Hello", "", nil)
		require.NoError(t, msg.Validate())
		require.Nil(t, msg.Data())
	})

	t.Run("empty title", func(t *testing.T) {
		t.Parallel()
		err := NewPushMessage("  ", "World", nil).Validate()
		require.ErrorIs(t, err, ErrPushTitleRequired)
	})

	t.Run("data is cloned", func(t *testing.T) {
		t.Parallel()
		src := map[string]string{"url": "/inbox"}
		msg := NewPushMessage("Hello", "World", src)
		src["url"] = "/mutated"
		require.Equal(t, "/inbox", msg.Data()["url"])
		cloned := msg.Data()
		cloned["url"] = "/again"
		require.Equal(t, "/inbox", msg.Data()["url"])
	})
}

func TestPushSendResult(t *testing.T) {
	t.Parallel()

	delivered := PushSendResult{
		StatusCode: 201,
		Expired:    false,
	}
	require.Equal(t, 201, delivered.StatusCode)
	require.False(t, delivered.Expired)

	expired := PushSendResult{
		StatusCode: 410,
		Expired:    true,
	}
	require.Equal(t, 410, expired.StatusCode)
	require.True(t, expired.Expired)
}
