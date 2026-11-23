package client

import (
	"context"
	"encoding/json"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestNewBrokerClientRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		cfg     *BrokerConnectionConfig
		wantErr string
	}{
		{
			name:    "nil config",
			wantErr: "config is required",
		},
		{
			name:    "empty config",
			cfg:     &BrokerConnectionConfig{},
			wantErr: "DSN is required",
		},
		{
			name: "missing api key",
			cfg: &BrokerConnectionConfig{
				DSN: "amqp://guest:guest@localhost:5672/",
			},
			wantErr: "api key is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewBrokerClient(ctx, tc.cfg)
			require.Error(t, err)
			require.ErrorContains(t, err, "validate config")
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestMarshalPushSend(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7().String()

	tests := []struct {
		name           string
		apiKey         string
		subscriptionID string
		msg            PushMessage
		wantErr        string
		wantJSON       string
	}{
		{
			name:           "valid with data",
			apiKey:         "ntf_test-key",
			subscriptionID: id,
			msg: PushMessage{
				Title: "Hello",
				Body:  "World",
				Data:  map[string]string{"url": "/inbox"},
			},
			wantJSON: `{"subscription_id":"` + id + `","api_key":"ntf_test-key","title":"Hello","body":"World","data":{"url":"/inbox"}}`,
		},
		{
			name:           "omits empty data",
			apiKey:         "  ntf_test-key  ",
			subscriptionID: "  " + id + "  ",
			msg: PushMessage{
				Title: "  Hello  ",
				Body:  "  World  ",
			},
			wantJSON: `{"subscription_id":"` + id + `","api_key":"ntf_test-key","title":"Hello","body":"World"}`,
		},
		{
			name:           "empty body allowed",
			apiKey:         "ntf_test-key",
			subscriptionID: id,
			msg:            PushMessage{Title: "Hello"},
			wantJSON:       `{"subscription_id":"` + id + `","api_key":"ntf_test-key","title":"Hello","body":""}`,
		},
		{
			name:           "missing subscription id",
			apiKey:         "ntf_test-key",
			msg:            PushMessage{Title: "Hello"},
			wantErr:        "subscription id is required",
			subscriptionID: "   ",
		},
		{
			name:           "invalid subscription id",
			apiKey:         "ntf_test-key",
			subscriptionID: "not-a-uuid",
			msg:            PushMessage{Title: "Hello"},
			wantErr:        "subscription id",
		},
		{
			name:           "missing title",
			apiKey:         "ntf_test-key",
			subscriptionID: id,
			msg:            PushMessage{Body: "World"},
			wantErr:        "title is required",
		},
		{
			name:           "missing api key",
			subscriptionID: id,
			msg:            PushMessage{Title: "Hello"},
			wantErr:        "api key is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, err := marshalPushSend(tc.apiKey, tc.subscriptionID, tc.msg)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.JSONEq(t, tc.wantJSON, string(body))

			var dto pushSendCommand
			require.NoError(t, json.Unmarshal(body, &dto))
			require.Equal(t, id, dto.SubscriptionID)
			require.Equal(t, "ntf_test-key", dto.APIKey)
			require.Equal(t, "Hello", dto.Title)
		})
	}
}

func TestSendRejectsClosedClient(t *testing.T) {
	t.Parallel()

	client := &BrokerClient{apiKey: "ntf_test-key"}
	client.closed.Store(true)

	err := client.Send(context.Background(), uuid.NewV7().String(), PushMessage{Title: "Hello"})
	require.ErrorContains(t, err, "broker client is closed")
}
