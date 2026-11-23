package client

import (
	"context"
	"fmt"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
)

type ClientRegistryClient struct {
	api notifyv1.ClientRegistryClient
}

func (c *grpcClient) ClientRegistryClient() *ClientRegistryClient {
	return &ClientRegistryClient{
		api: notifyv1.NewClientRegistryClient(c.conn),
	}
}

// Service is a caller-service api card.
type Service struct {
	ID           string
	Name         string
	Description  string
	MaskedAPIKey string
	CreatedAt    int64
	UpdatedAt    int64
}

// CreatedService is returned when a api client (or its API key) is created.
// Key is the plaintext API key and is only available once.
type CreatedService struct {
	Service
	Key string
}

// CreateService registers a caller service and returns the plaintext API key once.
func (c *ClientRegistryClient) CreateService(
	ctx context.Context,
	name string,
	description string,
) (*CreatedService, error) {
	resp, err := c.api.CreateClient(ctx, &notifyv1.CreateClientRequest{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	out := createdServiceFromProto(resp.GetClient(), resp.GetKey())
	return &out, nil
}

// GetService returns a caller service by id.
func (c *ClientRegistryClient) GetService(ctx context.Context, id string) (*Service, error) {
	resp, err := c.api.GetClient(ctx, &notifyv1.GetClientRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get service: %w", err)
	}

	svc := serviceFromProto(resp)
	return &svc, nil
}

// ListServices returns all caller services.
func (c *ClientRegistryClient) ListServices(ctx context.Context) ([]Service, error) {
	resp, err := c.api.ListClients(ctx, &notifyv1.ListClientsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}

	items := resp.GetServices()
	out := make([]Service, 0, len(items))
	for _, item := range items {
		out = append(out, serviceFromProto(item))
	}
	return out, nil
}

// DeleteService removes a caller service by id.
func (c *ClientRegistryClient) DeleteService(ctx context.Context, id string) error {
	_, err := c.api.DeleteClient(ctx, &notifyv1.DeleteClientRequest{Id: id})
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

// RotateAPIKey issues a new API key for the caller service. The plaintext key
// is returned once.
func (c *ClientRegistryClient) RotateAPIKey(ctx context.Context, id string) (*CreatedService, error) {
	resp, err := c.api.RotateAPIKey(ctx, &notifyv1.RotateAPIKeyRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("rotate api key: %w", err)
	}

	out := createdServiceFromProto(resp.GetService(), resp.GetKey())
	return &out, nil
}

func serviceFromProto(in *notifyv1.Client) Service {
	if in == nil {
		return Service{}
	}
	return Service{
		ID:           in.GetId(),
		Name:         in.GetName(),
		Description:  in.GetDescription(),
		MaskedAPIKey: in.GetMaskedApiKey(),
		CreatedAt:    in.GetCreatedAt(),
		UpdatedAt:    in.GetUpdatedAt(),
	}
}

func createdServiceFromProto(in *notifyv1.Client, key string) CreatedService {
	return CreatedService{
		Service: serviceFromProto(in),
		Key:     key,
	}
}
