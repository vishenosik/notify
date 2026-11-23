package client

import (
	"context"
	"fmt"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
)

type WebPushClient struct {
	api notifyv1.WebPushClient
}

func (c *grpcClient) WebPushClient() *WebPushClient {
	return &WebPushClient{
		api: notifyv1.NewWebPushClient(c.conn),
	}
}

// GetVAPIDPublicKey returns the server VAPID public key for Web Push subscriptions.
func (c *WebPushClient) GetVAPIDPublicKey(ctx context.Context) (string, error) {
	resp, err := c.api.GetVapidPublicKey(ctx, &notifyv1.GetVapidPublicKeyRequest{})
	if err != nil {
		return "", fmt.Errorf("get vapid public key: %w", err)
	}
	return resp.GetPublicKey(), nil
}

// SubscribeInput is the browser Web Push subscription to register.
type SubscribeInput struct {
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent string
}

// Subscribe registers a browser Web Push subscription and returns its id.
func (c *WebPushClient) Subscribe(ctx context.Context, in SubscribeInput) (string, error) {
	resp, err := c.api.Subscribe(ctx, &notifyv1.SubscribeRequest{
		Endpoint:  in.Endpoint,
		P256Dh:    in.P256dh,
		Auth:      in.Auth,
		UserAgent: in.UserAgent,
	})
	if err != nil {
		return "", fmt.Errorf("subscribe: %w", err)
	}
	return resp.GetId(), nil
}

// Unsubscribe removes a previously registered Web Push subscription.
func (c *WebPushClient) Unsubscribe(ctx context.Context, subscriptionID string) error {
	_, err := c.api.Unsubscribe(ctx, &notifyv1.UnsubscribeRequest{Id: subscriptionID})
	if err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}
	return nil
}
