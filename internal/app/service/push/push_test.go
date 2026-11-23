package push

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vishenosik/notify/internal/app/command"
	ifacemock "github.com/vishenosik/notify/internal/app/interface/mock"
	"github.com/vishenosik/notify/internal/domain/entity"
	repomock "github.com/vishenosik/notify/internal/domain/repository/mock"
	"github.com/vishenosik/notify/internal/domain/xerrors"
	"github.com/vishenosik/notify/pkg/rabbitmq"
)

func testSubscription(clientID uuid.UUID) *entity.PushSubscription {
	return entity.NewPushSubscription(entity.PushSubscriptionVariants{
		ClientID: clientID,
		Endpoint: "https://push.example/sub",
		P256dh:   "p256dh-key",
		Auth:     "auth-secret",
	})
}

func TestService_VAPIDPublicKey(t *testing.T) {
	t.Parallel()

	t.Run("configured", func(t *testing.T) {
		t.Parallel()
		got, err := New(nil, nil, nil, "public", "private").VAPIDPublicKey()
		require.NoError(t, err)
		require.Equal(t, "public", got.PublicKey)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		got, err := New(nil, nil, nil, "", "").VAPIDPublicKey()
		require.ErrorIs(t, err, xerrors.ErrValidation)
		require.Nil(t, got)
	})
}

func TestService_Subscribe(t *testing.T) {
	t.Parallel()

	clientID := uuid.NewV7()
	cmd := &command.SubscribePushCommand{
		Endpoint:  "https://push.example/sub",
		P256dh:    "p256dh-key",
		Auth:      "auth-secret",
		UserAgent: "Mozilla/5.0",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		notifier := repomock.NewMockWebPushNotifier(t)
		subs.EXPECT().Upsert(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, sub entity.ValidatedPushSubscription) (*entity.PushSubscription, error) {
				require.Equal(t, clientID, sub.Value.ClientID())
				return sub.Value, nil
			}).Once()

		got, err := New(subs, keys, notifier, "pub", "priv").Subscribe(context.Background(), clientID, cmd)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil(), got.ID)
	})

	t.Run("validation", func(t *testing.T) {
		t.Parallel()
		got, err := New(
			repomock.NewMockPushSubscriptionRepository(t),
			ifacemock.NewMockAPIKeyUseCase(t),
			repomock.NewMockWebPushNotifier(t),
			"pub",
			"priv",
		).Subscribe(context.Background(), clientID, &command.SubscribePushCommand{})
		require.ErrorIs(t, err, xerrors.ErrValidation)
		require.Nil(t, got)
	})
}

func TestService_Unsubscribe(t *testing.T) {
	t.Parallel()

	clientID := uuid.NewV7()
	sub := testSubscription(clientID)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		subs.EXPECT().DeleteByID(mock.Anything, sub.ID()).Return(nil).Once()
		require.NoError(t, New(subs, ifacemock.NewMockAPIKeyUseCase(t), repomock.NewMockWebPushNotifier(t), "", "").
			Unsubscribe(context.Background(), clientID, sub.ID()))
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		id := uuid.NewV7()
		subs.EXPECT().GetByID(mock.Anything, id).Return(nil, xerrors.ErrNotFound).Once()
		err := New(subs, ifacemock.NewMockAPIKeyUseCase(t), repomock.NewMockWebPushNotifier(t), "", "").
			Unsubscribe(context.Background(), clientID, id)
		require.ErrorIs(t, err, xerrors.ErrNotFound)
	})

	t.Run("client mismatch", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		err := New(subs, ifacemock.NewMockAPIKeyUseCase(t), repomock.NewMockWebPushNotifier(t), "", "").
			Unsubscribe(context.Background(), uuid.NewV7(), sub.ID())
		require.ErrorIs(t, err, xerrors.ErrNotFound)
	})
}

func TestService_Deliver(t *testing.T) {
	t.Parallel()

	clientID := uuid.NewV7()
	sub := testSubscription(clientID)
	cmd, err := entity.NewValidated(entity.NewPushCommand(sub.ID(), "ntf_secret", entity.NewPushMessage("Hello", "World", nil)))
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		notifier := repomock.NewMockWebPushNotifier(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(clientID, nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		notifier.EXPECT().Send(mock.Anything, sub, mock.Anything).Return(&entity.PushSendResult{StatusCode: 201}, nil).Once()

		require.NoError(t, New(subs, keys, notifier, "pub", "priv").Deliver(context.Background(), cmd))
	})

	t.Run("auth fail drops", func(t *testing.T) {
		t.Parallel()
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(uuid.Nil(), xerrors.ErrAuth).Once()
		err := New(repomock.NewMockPushSubscriptionRepository(t), keys, repomock.NewMockWebPushNotifier(t), "pub", "priv").
			Deliver(context.Background(), cmd)
		require.ErrorIs(t, err, rabbitmq.ErrDropMessage)
	})

	t.Run("missing subscription drops", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(clientID, nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(nil, xerrors.ErrNotFound).Once()
		err := New(subs, keys, repomock.NewMockWebPushNotifier(t), "pub", "priv").Deliver(context.Background(), cmd)
		require.ErrorIs(t, err, rabbitmq.ErrDropMessage)
	})

	t.Run("client mismatch drops", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(uuid.NewV7(), nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		err := New(subs, keys, repomock.NewMockWebPushNotifier(t), "pub", "priv").Deliver(context.Background(), cmd)
		require.ErrorIs(t, err, rabbitmq.ErrDropMessage)
	})

	t.Run("missing vapid requeues", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(clientID, nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		err := New(subs, keys, repomock.NewMockWebPushNotifier(t), "", "").Deliver(context.Background(), cmd)
		require.ErrorIs(t, err, xerrors.VAPIDNotConfiguredError())
		require.NotErrorIs(t, err, rabbitmq.ErrDropMessage)
	})

	t.Run("whitespace vapid requeues", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(clientID, nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		err := New(subs, keys, repomock.NewMockWebPushNotifier(t), "   ", "   ").Deliver(context.Background(), cmd)
		require.ErrorIs(t, err, xerrors.VAPIDNotConfiguredError())
		require.NotErrorIs(t, err, rabbitmq.ErrDropMessage)
	})

	t.Run("expired deletes subscription", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		notifier := repomock.NewMockWebPushNotifier(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(clientID, nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		notifier.EXPECT().Send(mock.Anything, sub, mock.Anything).Return(&entity.PushSendResult{Expired: true, StatusCode: 410}, nil).Once()
		subs.EXPECT().DeleteByID(mock.Anything, sub.ID()).Return(nil).Once()
		require.NoError(t, New(subs, keys, notifier, "pub", "priv").Deliver(context.Background(), cmd))
	})

	t.Run("expired delete not found ignored", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		notifier := repomock.NewMockWebPushNotifier(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(clientID, nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		notifier.EXPECT().Send(mock.Anything, sub, mock.Anything).Return(&entity.PushSendResult{Expired: true}, nil).Once()
		subs.EXPECT().DeleteByID(mock.Anything, sub.ID()).Return(xerrors.ErrNotFound).Once()
		require.NoError(t, New(subs, keys, notifier, "pub", "priv").Deliver(context.Background(), cmd))
	})

	t.Run("send error requeues", func(t *testing.T) {
		t.Parallel()
		subs := repomock.NewMockPushSubscriptionRepository(t)
		keys := ifacemock.NewMockAPIKeyUseCase(t)
		notifier := repomock.NewMockWebPushNotifier(t)
		keys.EXPECT().Authenticate(mock.Anything, "ntf_secret").Return(clientID, nil).Once()
		subs.EXPECT().GetByID(mock.Anything, sub.ID()).Return(sub, nil).Once()
		notifier.EXPECT().Send(mock.Anything, sub, mock.Anything).Return(nil, errors.New("provider down")).Once()
		err := New(subs, keys, notifier, "pub", "priv").Deliver(context.Background(), cmd)
		require.ErrorContains(t, err, "provider down")
		require.NotErrorIs(t, err, rabbitmq.ErrDropMessage)
	})
}
