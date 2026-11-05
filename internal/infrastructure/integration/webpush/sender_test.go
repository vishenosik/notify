package webpush

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
)

var _ repository.WebPushNotifier = (*Sender)(nil)

type stubClient struct {
	status   int
	body     string
	err      error
	closeErr error
	readErr  error
	got      *http.Request
}

func (s *stubClient) Do(req *http.Request) (*http.Response, error) {
	s.got = req
	if s.err != nil {
		return nil, s.err
	}

	var body io.ReadCloser
	switch {
	case s.readErr != nil:
		body = failCloser{readErr: s.readErr}
	case s.closeErr != nil:
		body = failCloser{Reader: strings.NewReader(s.body), closeErr: s.closeErr}
	default:
		body = io.NopCloser(strings.NewReader(s.body))
	}

	return &http.Response{
		StatusCode: s.status,
		Body:       body,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

type failCloser struct {
	io.Reader
	readErr  error
	closeErr error
}

func (c failCloser) Read(p []byte) (int, error) {
	if c.readErr != nil {
		return 0, c.readErr
	}
	if c.Reader == nil {
		return 0, io.EOF
	}
	return c.Reader.Read(p)
}

func (c failCloser) Close() error {
	return c.closeErr
}

func testSender(t *testing.T, client HTTPClient) *Sender {
	t.Helper()
	priv, pub := generateVAPIDPair(t)
	sender, err := NewSender(&Config{
		VAPIDPublicKey:  pub,
		VAPIDPrivateKey: priv,
		VAPIDSubject:    "ops@example.com",
		TTL:             30,
		HTTPClient:      client,
	})
	require.NoError(t, err)
	return sender
}

func testSubscription(t *testing.T) *entity.PushSubscription {
	t.Helper()
	p256dh, auth := testSubscriptionKeys(t)
	return entity.NewPushSubscription(entity.PushSubscriptionVariants{
		Endpoint:  "https://push.example/sub",
		P256dh:    p256dh,
		Auth:      auth,
		UserAgent: "Mozilla/5.0",
	})
}

func testMessage() *entity.PushMessage {
	msg := entity.NewPushMessage("Hello", "World", map[string]string{"url": "/inbox"})
	return &msg
}

func testSubscriptionKeys(t *testing.T) (p256dh, auth string) {
	t.Helper()

	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	require.NoError(t, err)

	secret := make([]byte, 16)
	_, err = rand.Read(secret)
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes()),
		base64.RawURLEncoding.EncodeToString(secret)
}

func TestNewSender(t *testing.T) {
	t.Parallel()

	t.Run("rejects invalid config", func(t *testing.T) {
		t.Parallel()
		sender, err := NewSender(&Config{})
		require.Error(t, err)
		require.Nil(t, sender)
		require.ErrorContains(t, err, "webpush config validation failed")
	})

	t.Run("applies defaults", func(t *testing.T) {
		t.Parallel()
		priv, pub := generateVAPIDPair(t)
		cfg := &Config{
			VAPIDPublicKey:  pub,
			VAPIDPrivateKey: priv,
			VAPIDSubject:    "ops@example.com",
		}
		sender, err := NewSender(cfg)
		require.NoError(t, err)
		require.Equal(t, pub, sender.publicKey)
		require.Equal(t, priv, sender.privateKey)
		require.Equal(t, "ops@example.com", sender.subject)
		require.Equal(t, uint16(defaultTTL), sender.ttl)
		require.NotNil(t, sender.client)
	})
}

func TestSenderSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		client      stubClient
		wantCode    int
		wantExpired bool
		wantErr     string
	}{
		{
			name:     "created",
			client:   stubClient{status: http.StatusCreated},
			wantCode: http.StatusCreated,
		},
		{
			name:     "ok",
			client:   stubClient{status: http.StatusOK},
			wantCode: http.StatusOK,
		},
		{
			name:        "gone",
			client:      stubClient{status: http.StatusGone},
			wantCode:    http.StatusGone,
			wantExpired: true,
		},
		{
			name:        "not found",
			client:      stubClient{status: http.StatusNotFound},
			wantCode:    http.StatusNotFound,
			wantExpired: true,
		},
		{
			name:        "forbidden",
			client:      stubClient{status: http.StatusForbidden, body: "BadJwtToken"},
			wantCode:    http.StatusForbidden,
			wantExpired: true,
		},
		{
			name:     "provider error with body",
			client:   stubClient{status: http.StatusBadRequest, body: "invalid payload"},
			wantCode: http.StatusBadRequest,
			wantErr:  "push provider status 400: invalid payload",
		},
		{
			name:     "provider error empty body",
			client:   stubClient{status: http.StatusInternalServerError},
			wantCode: http.StatusInternalServerError,
			wantErr:  "push provider status 500",
		},
		{
			name:    "transport error",
			client:  stubClient{err: context.DeadlineExceeded},
			wantErr: "send notification",
		},
		{
			name:    "close error",
			client:  stubClient{status: http.StatusCreated, closeErr: errors.New("close failed")},
			wantErr: "close push response",
		},
		{
			name:     "body read error ignored",
			client:   stubClient{status: http.StatusCreated, readErr: errors.New("read failed")},
			wantCode: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := tc.client
			sender := testSender(t, &client)

			got, err := sender.Send(t.Context(), testSubscription(t), testMessage())
			if tc.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantCode, got.StatusCode)
			require.Equal(t, tc.wantExpired, got.Expired)
			require.Equal(t, http.MethodPost, client.got.Method)
			require.Equal(t, "https://push.example/sub", client.got.URL.String())
			require.Equal(t, "30", client.got.Header.Get("TTL"))
		})
	}
}

func TestSenderSendProviderErrorKeepsStatus(t *testing.T) {
	t.Parallel()

	client := &stubClient{status: http.StatusBadRequest, body: "invalid payload"}
	sender := testSender(t, client)
	sub := testSubscription(t)
	msg := testMessage()

	got, err := sender.send(t.Context(), pushTarget{
		endpoint: sub.Endpoint(),
		p256dh:   sub.P256dh(),
		auth:     sub.Auth(),
	}, pushMessage{
		Title: msg.Title(),
		Body:  msg.Body(),
		Data:  msg.Data(),
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "push provider status 400: invalid payload")
	require.Equal(t, http.StatusBadRequest, got.statusCode)
	require.False(t, got.expired)
}
