package entity

import (
	"fmt"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestPushCommandValidate(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	msg := NewPushMessage("Hello", "World", map[string]string{"url": "/inbox"})

	tests := []struct {
		name    string
		cmd     *PushCommand
		wantErr error
	}{
		{
			name: "valid",
			cmd:  NewPushCommand(id, "ntf_test-key", msg),
		},
		{
			name: "empty body allowed",
			cmd:  NewPushCommand(id, "ntf_test-key", NewPushMessage("Hello", "", nil)),
		},
		{
			name:    "nil subscription id",
			cmd:     NewPushCommand(uuid.Nil(), "ntf_test-key", msg),
			wantErr: ErrPushSubscriptionIDRequired,
		},
		{
			name:    "empty api key",
			cmd:     NewPushCommand(id, "", msg),
			wantErr: ErrPushAPIKeyRequired,
		},
		{
			name:    "whitespace api key",
			cmd:     NewPushCommand(id, "  \t", msg),
			wantErr: ErrPushAPIKeyRequired,
		},
		{
			name:    "empty title",
			cmd:     NewPushCommand(id, "ntf_test-key", NewPushMessage("  ", "World", nil)),
			wantErr: ErrPushTitleRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.cmd.Validate()
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, id, tc.cmd.SubscriptionID())
			require.Equal(t, "ntf_test-key", tc.cmd.APIKey())
			require.Equal(t, "Hello", tc.cmd.Message().Title())
		})
	}
}

func TestPushCommandValidateCollectsErrors(t *testing.T) {
	t.Parallel()

	err := NewPushCommand(uuid.Nil(), "", NewPushMessage("", "", nil)).Validate()
	require.ErrorIs(t, err, ErrPushSubscriptionIDRequired)
	require.ErrorIs(t, err, ErrPushAPIKeyRequired)
	require.ErrorIs(t, err, ErrPushTitleRequired)
}

func TestPushCommandTrimsAPIKey(t *testing.T) {
	t.Parallel()

	cmd := NewPushCommand(uuid.NewV7(), "  ntf_test-key  ", NewPushMessage("Hello", "", nil))
	require.Equal(t, "ntf_test-key", cmd.APIKey())
	require.NoError(t, cmd.Validate())
}

func TestPushCommandClonesData(t *testing.T) {
	t.Parallel()

	src := map[string]string{"url": "/inbox"}
	msg := NewPushMessage("Hello", "World", src)
	cmd := NewPushCommand(uuid.NewV7(), "ntf_test-key", msg)

	src["url"] = "/mutated"
	require.Equal(t, "/inbox", cmd.Message().Data()["url"])

	cloned := cmd.Message().Data()
	cloned["url"] = "/again"
	require.Equal(t, "/inbox", cmd.Message().Data()["url"])
}

func TestPushCommandDoesNotExposeAPIKey(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	secret := "ntf_super-secret-key"
	cmd := NewPushCommand(id, secret, NewPushMessage("Hello", "World", nil))

	require.NotContains(t, cmd.String(), secret)
	require.NotContains(t, cmd.GoString(), secret)
	require.NotContains(t, fmt.Sprintf("%v", cmd), secret)
	require.NotContains(t, fmt.Sprintf("%+v", cmd), secret)
	require.NotContains(t, fmt.Sprintf("%#v", cmd), secret)
	require.NotContains(t, fmt.Sprintf("%v", *cmd), secret)
	require.Contains(t, cmd.String(), id.String())
}
