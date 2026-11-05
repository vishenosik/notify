package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
)

var _ repository.Consumer = (*Client)(nil)

type stubTransport struct {
	deliveries <-chan delivery
	consumeErr error
	closeErr   error
	closed     bool
}

func (s *stubTransport) Consume(_ context.Context) (<-chan delivery, error) {
	if s.consumeErr != nil {
		return nil, s.consumeErr
	}
	return s.deliveries, nil
}

func (s *stubTransport) Close() error {
	s.closed = true
	return s.closeErr
}

func testCommand(t *testing.T) *entity.PushCommand {
	t.Helper()
	return entity.NewPushCommand(
		uuid.NewV7(),
		"ntf_test-key",
		entity.NewPushMessage("Hello", "World", map[string]string{"url": "/inbox"}),
	)
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("rejects invalid config", func(t *testing.T) {
		t.Parallel()
		client, err := New(&Config{})
		require.Error(t, err)
		require.Nil(t, client)
		require.ErrorContains(t, err, "rabbitmq config validation failed")
		require.ErrorContains(t, err, "URL is required")
	})
}

func TestClientClose(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		tr := &stubTransport{}
		client := newClient(tr)
		require.NoError(t, client.Close())
		require.True(t, tr.closed)
	})

	t.Run("wraps error", func(t *testing.T) {
		t.Parallel()
		tr := &stubTransport{closeErr: errors.New("close failed")}
		client := newClient(tr)
		err := client.Close()
		require.ErrorContains(t, err, "close rabbitmq")
		require.ErrorContains(t, err, "close failed")
		require.True(t, tr.closed)
	})
}

func TestClientConsume(t *testing.T) {
	t.Parallel()

	valid := testCommand(t)
	validBody := commandJSON(t, valid)

	tests := []struct {
		name        string
		body        []byte
		handlerErr  error
		ackErr      error
		nackErr     error
		wantAck     bool
		wantNack    bool
		wantRequeue bool
		wantErr     string
	}{
		{
			name:    "acks on handler nil",
			body:    validBody,
			wantAck: true,
			wantErr: "broker consumer closed",
		},
		{
			name:        "nacks drop on bad json",
			body:        []byte("{"),
			wantNack:    true,
			wantRequeue: false,
			wantErr:     "broker consumer closed",
		},
		{
			name:        "nacks drop on invalid command",
			body:        []byte(`{"subscription_id":"` + uuid.NewV7().String() + `","api_key":"ntf_x","title":""}`),
			wantNack:    true,
			wantRequeue: false,
			wantErr:     "broker consumer closed",
		},
		{
			name:        "nacks drop on ErrDrop",
			body:        validBody,
			handlerErr:  repository.ErrDrop,
			wantNack:    true,
			wantRequeue: false,
			wantErr:     "broker consumer closed",
		},
		{
			name:        "nacks drop on wrapped ErrDrop",
			body:        validBody,
			handlerErr:  fmtWrap(repository.ErrDrop),
			wantNack:    true,
			wantRequeue: false,
			wantErr:     "broker consumer closed",
		},
		{
			name:        "nacks requeue on other handler errors",
			body:        validBody,
			handlerErr:  errors.New("transient"),
			wantNack:    true,
			wantRequeue: true,
			wantErr:     "broker consumer closed",
		},
		{
			name:    "wraps ack errors",
			body:    validBody,
			ackErr:  errors.New("ack failed"),
			wantAck: true,
			wantErr: "ack message",
		},
		{
			name:        "wraps nack drop errors",
			body:        []byte("{"),
			nackErr:     errors.New("nack failed"),
			wantNack:    true,
			wantRequeue: false,
			wantErr:     "nack drop invalid message",
		},
		{
			name:        "wraps nack drop errors on ErrDrop",
			body:        validBody,
			handlerErr:  repository.ErrDrop,
			nackErr:     errors.New("nack failed"),
			wantNack:    true,
			wantRequeue: false,
			wantErr:     "nack drop message",
		},
		{
			name:        "wraps nack requeue errors",
			body:        validBody,
			handlerErr:  errors.New("transient"),
			nackErr:     errors.New("nack failed"),
			wantNack:    true,
			wantRequeue: true,
			wantErr:     "nack requeue message",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var acked bool
			var nacked bool
			var requeue bool
			deliveries := make(chan delivery, 1)
			deliveries <- delivery{
				Body: tc.body,
				Ack: func() error {
					acked = true
					return tc.ackErr
				},
				Nack: func(rq bool) error {
					nacked = true
					requeue = rq
					return tc.nackErr
				},
			}
			close(deliveries)

			client := newClient(&stubTransport{deliveries: deliveries})
			err := client.Consume(t.Context(), func(context.Context, *entity.PushCommand) error {
				return tc.handlerErr
			})
			require.ErrorContains(t, err, tc.wantErr)
			if tc.wantErr != "broker consumer closed" {
				require.ErrorContains(t, err, errorSuffix(tc.ackErr, tc.nackErr))
			}
			require.Equal(t, tc.wantAck, acked)
			require.Equal(t, tc.wantNack, nacked)
			if tc.wantNack {
				require.Equal(t, tc.wantRequeue, requeue)
			}
		})
	}

	t.Run("passes decoded command to handler", func(t *testing.T) {
		t.Parallel()
		deliveries := make(chan delivery, 1)
		deliveries <- delivery{
			Body: validBody,
			Ack:  func() error { return nil },
			Nack: func(bool) error { return errors.New("unexpected nack") },
		}
		close(deliveries)

		var got *entity.PushCommand
		client := newClient(&stubTransport{deliveries: deliveries})
		err := client.Consume(t.Context(), func(_ context.Context, command *entity.PushCommand) error {
			got = command
			return nil
		})
		require.ErrorContains(t, err, "broker consumer closed")
		require.NotNil(t, got)
		require.Equal(t, valid.SubscriptionID(), got.SubscriptionID())
		require.Equal(t, valid.APIKey(), got.APIKey())
		require.Equal(t, valid.Message().Title(), got.Message().Title())
		require.Equal(t, valid.Message().Body(), got.Message().Body())
		require.Equal(t, valid.Message().Data(), got.Message().Data())
	})

	t.Run("wraps consume start errors", func(t *testing.T) {
		t.Parallel()
		client := newClient(&stubTransport{consumeErr: errors.New("consume failed")})
		err := client.Consume(t.Context(), func(context.Context, *entity.PushCommand) error {
			return nil
		})
		require.ErrorContains(t, err, "start consume")
		require.ErrorContains(t, err, "consume failed")
	})

	t.Run("stops on ctx cancel", func(t *testing.T) {
		t.Parallel()
		client := newClient(&stubTransport{deliveries: make(chan delivery)})
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() {
			done <- client.Consume(ctx, func(context.Context, *entity.PushCommand) error {
				return nil
			})
		}()
		cancel()

		select {
		case err := <-done:
			require.ErrorIs(t, err, context.Canceled)
		case <-t.Context().Done():
			t.Fatal("Consume did not return after cancel")
		}
	})
}

func commandJSON(t *testing.T, command *entity.PushCommand) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"subscription_id": command.SubscriptionID().String(),
		"api_key":         command.APIKey(),
		"title":           command.Message().Title(),
		"body":            command.Message().Body(),
		"data":            command.Message().Data(),
	})
	require.NoError(t, err)
	return body
}

func fmtWrap(err error) error {
	return fmt.Errorf("unauthorized: %w", err)
}

func errorSuffix(ackErr, nackErr error) string {
	if ackErr != nil {
		return ackErr.Error()
	}
	if nackErr != nil {
		return nackErr.Error()
	}
	return ""
}
