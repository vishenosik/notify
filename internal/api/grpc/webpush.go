package grpcapi

import (
	"context"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	"github.com/vishenosik/notify/internal/api/grpc/mapper"
	middleware "github.com/vishenosik/notify/internal/api/grpc/middleware"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

type webPushServer struct {
	notifyv1.UnimplementedWebPushServer
	useCase interfaces.PushUseCase
}

func newWebPushServer(useCase interfaces.PushUseCase) notifyv1.WebPushServer {
	return &webPushServer{useCase: useCase}
}

func (s *webPushServer) clientID(ctx context.Context) (uuid.UUID, error) {
	id, ok := middleware.ClientIDFromContext(ctx)
	if !ok {
		return uuid.Nil(), status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return id, nil
}

func (s *webPushServer) GetVapidPublicKey(
	ctx context.Context,
	_ *notifyv1.GetVapidPublicKeyRequest,
) (*notifyv1.GetVapidPublicKeyResponse, error) {
	if _, err := s.clientID(ctx); err != nil {
		return nil, err
	}

	result, err := s.useCase.VAPIDPublicKey()
	if err != nil {
		return nil, err
	}

	return mapper.VAPIDPublicKeyResponse(result), nil
}

func (s *webPushServer) Subscribe(
	ctx context.Context,
	req *notifyv1.SubscribeRequest,
) (*notifyv1.SubscribeResponse, error) {
	clientID, err := s.clientID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.useCase.Subscribe(ctx, clientID, mapper.SubscribePushCommand(req))
	if err != nil {
		return nil, err
	}

	return mapper.SubscribeResponse(result), nil
}

func (s *webPushServer) Unsubscribe(
	ctx context.Context,
	req *notifyv1.UnsubscribeRequest,
) (*notifyv1.UnsubscribeResponse, error) {
	clientID, err := s.clientID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, xerrors.InvalidIDError()
	}

	if err := s.useCase.Unsubscribe(ctx, clientID, id); err != nil {
		return nil, err
	}

	return &notifyv1.UnsubscribeResponse{}, nil
}

var _ notifyv1.WebPushServer = (*webPushServer)(nil)
