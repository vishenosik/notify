package grpcapi

import (
	"context"
	"uuid"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	"github.com/vishenosik/notify/internal/api/grpc/mapper"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

type clientRegistryServer struct {
	notifyv1.UnimplementedClientRegistryServer
	clients interfaces.ClientUseCase
	keys    interfaces.APIKeyUseCase
}

func newClientRegistryServer(
	clients interfaces.ClientUseCase,
	keys interfaces.APIKeyUseCase,
) notifyv1.ClientRegistryServer {
	return &clientRegistryServer{
		clients: clients,
		keys:    keys,
	}
}

func (s *clientRegistryServer) CreateClient(
	ctx context.Context,
	req *notifyv1.CreateClientRequest,
) (*notifyv1.CreateClientResponse, error) {
	result, err := s.clients.Create(ctx, mapper.CreateClientCommand(req))
	if err != nil {
		return nil, err
	}

	return mapper.CreateClientResponse(result), nil
}

func (s *clientRegistryServer) GetClient(
	ctx context.Context,
	req *notifyv1.GetClientRequest,
) (*notifyv1.Client, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, xerrors.InvalidIDError()
	}

	result, err := s.clients.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.Client(result), nil
}

func (s *clientRegistryServer) DeleteClient(
	ctx context.Context,
	req *notifyv1.DeleteClientRequest,
) (*notifyv1.DeleteClientResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, xerrors.InvalidIDError()
	}

	if err := s.clients.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &notifyv1.DeleteClientResponse{}, nil
}

func (s *clientRegistryServer) ListClients(
	ctx context.Context,
	_ *notifyv1.ListClientsRequest,
) (*notifyv1.ListClientsResponse, error) {
	results, err := s.clients.List(ctx)
	if err != nil {
		return nil, err
	}

	return &notifyv1.ListClientsResponse{Services: mapper.Clients(results)}, nil
}

func (s *clientRegistryServer) RotateAPIKey(
	ctx context.Context,
	req *notifyv1.RotateAPIKeyRequest,
) (*notifyv1.RotateAPIKeyResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, xerrors.InvalidIDError()
	}

	result, err := s.keys.Rotate(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.RotateAPIKeyResponse(result), nil
}

var _ notifyv1.ClientRegistryServer = (*clientRegistryServer)(nil)
