package webpush

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	webpushlib "github.com/SherClockHolmes/webpush-go"

	"github.com/vishenosik/notify/internal/domain/entity"
)

type Sender struct {
	publicKey  string
	privateKey string
	subject    string
	ttl        uint16
	client     HTTPClient
}

// NewSender validates cfg and returns a sender that delivers one push per Send call.
func NewSender(cfg *Config) (*Sender, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("webpush config validation failed: %w", err)
	}

	sender := &Sender{
		publicKey:  cfg.VAPIDPublicKey,
		privateKey: cfg.VAPIDPrivateKey,
		subject:    cfg.VAPIDSubject,
		ttl:        cfg.TTL,
		client:     cfg.HTTPClient,
	}

	return sender, nil
}

// Send encrypts and POSTs one notification. HTTP 403/404/410 are returned as
// Expired=true with a nil error so the caller can drop the subscription.
func (s *Sender) Send(ctx context.Context, target *entity.PushSubscription, message *entity.PushMessage) (*entity.PushSendResult, error) {
	result, err := s.send(ctx, pushTarget{
		endpoint: target.Endpoint(),
		p256dh:   target.P256dh(),
		auth:     target.Auth(),
	}, pushMessage{
		Title: message.Title(),
		Body:  message.Body(),
		Data:  message.Data(),
	})
	if err != nil {
		return nil, err
	}
	return &entity.PushSendResult{
		StatusCode: result.statusCode,
		Expired:    result.expired,
	}, nil
}

type pushTarget struct {
	endpoint string
	p256dh   string
	auth     string
}

type pushMessage struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"`
}

type pushSendResult struct {
	statusCode int
	expired    bool
}

// send encrypts and POSTs one notification. HTTP 403/404/410 are returned as
// Expired=true with a nil error so the caller can drop the subscription.
func (s *Sender) send(ctx context.Context, target pushTarget, message pushMessage) (*pushSendResult, error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal push payload: %w", err)
	}

	resp, err := webpushlib.SendNotificationWithContext(ctx, payload,
		&webpushlib.Subscription{
			Endpoint: target.endpoint,
			Keys: webpushlib.Keys{
				P256dh: target.p256dh,
				Auth:   target.auth,
			},
		},
		&webpushlib.Options{
			HTTPClient:      s.client,
			Subscriber:      s.subject,
			TTL:             int(s.ttl),
			VAPIDPublicKey:  s.publicKey,
			VAPIDPrivateKey: s.privateKey,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("send notification: %w", err)
	}

	body := readResponseBody(resp)
	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("close push response: %w", err)
	}

	result := &pushSendResult{
		statusCode: resp.StatusCode,
		expired:    isExpiredStatus(resp.StatusCode),
	}
	if result.expired {
		return result, nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return result, fmt.Errorf("push provider status %d%s", resp.StatusCode, body)
	}
	return result, nil
}

func readResponseBody(resp *http.Response) string {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil || len(body) == 0 {
		return ""
	}
	return ": " + strings.TrimSpace(string(body))
}

func isExpiredStatus(code int) bool {
	return code == http.StatusGone || code == http.StatusNotFound || code == http.StatusForbidden
}
