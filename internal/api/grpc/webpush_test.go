package grpcapi

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	middleware "github.com/vishenosik/notify/internal/api/grpc/middleware"
	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
	ifacemock "github.com/vishenosik/notify/internal/app/interface/mock"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func TestWebPushSubscribe(t *testing.T) {
	t.Parallel()

	clientID := uuid.NewV7()
	subID := uuid.NewV7()
	useCase := ifacemock.NewMockPushUseCase(t)
	useCase.EXPECT().Subscribe(mock.Anything, clientID, mock.Anything).
		RunAndReturn(func(_ context.Context, _ uuid.UUID, cmd *command.SubscribePushCommand) (*common.PushSubscriptionResult, error) {
			require.Equal(t, "https://push.example/sub", cmd.Endpoint)
			require.Equal(t, "p256dh", cmd.P256dh)
			return &common.PushSubscriptionResult{ID: subID}, nil
		}).Once()

	srv := newWebPushServer(useCase)
	got, err := srv.Subscribe(middleware.WithClientID(context.Background(), clientID), &notifyv1.SubscribeRequest{
		Endpoint:  "https://push.example/sub",
		P256Dh:    "p256dh",
		Auth:      "auth",
		UserAgent: "ua",
	})
	require.NoError(t, err)
	require.Equal(t, subID.String(), got.GetId())
}

func TestWebPushUnsubscribeInvalidID(t *testing.T) {
	t.Parallel()

	srv := newWebPushServer(ifacemock.NewMockPushUseCase(t))
	_, err := srv.Unsubscribe(middleware.WithClientID(context.Background(), uuid.NewV7()), &notifyv1.UnsubscribeRequest{Id: "not-a-uuid"})
	require.ErrorIs(t, err, xerrors.ErrValidation)
}

func TestWebPushUnsubscribeMissingClientID(t *testing.T) {
	t.Parallel()

	srv := newWebPushServer(ifacemock.NewMockPushUseCase(t))
	_, err := srv.Unsubscribe(context.Background(), &notifyv1.UnsubscribeRequest{Id: uuid.NewV7().String()})
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestWebPushGetVapidPublicKey(t *testing.T) {
	t.Parallel()

	useCase := ifacemock.NewMockPushUseCase(t)
	useCase.EXPECT().VAPIDPublicKey().Return(&common.VAPIDPublicKeyResult{PublicKey: "pub"}, nil).Once()
	srv := newWebPushServer(useCase)
	got, err := srv.GetVapidPublicKey(middleware.WithClientID(context.Background(), uuid.NewV7()), &notifyv1.GetVapidPublicKeyRequest{})
	require.NoError(t, err)
	require.Equal(t, "pub", got.GetPublicKey())
}
