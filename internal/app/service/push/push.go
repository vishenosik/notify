package push

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"uuid"

	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/app/mapper"
	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

type service struct {
	subscriptions repository.PushSubscriptionRepository
	keys          interfaces.APIKeyUseCase
	notifier      repository.WebPushNotifier
	vapidPublic   string
	vapidPrivate  string
}

func New(
	subscriptions repository.PushSubscriptionRepository,
	keys interfaces.APIKeyUseCase,
	notifier repository.WebPushNotifier,
	vapidPublic string,
	vapidPrivate string,
) interfaces.PushUseCase {
	return &service{
		subscriptions: subscriptions,
		keys:          keys,
		notifier:      notifier,
		vapidPublic:   strings.TrimSpace(vapidPublic),
		vapidPrivate:  strings.TrimSpace(vapidPrivate),
	}
}

func (s *service) VAPIDPublicKey() (*common.VAPIDPublicKeyResult, error) {
	if s.vapidPublic == "" {
		return nil, xerrors.VAPIDNotConfiguredError()
	}
	return &common.VAPIDPublicKeyResult{PublicKey: s.vapidPublic}, nil
}

func (s *service) Subscribe(
	ctx context.Context,
	clientID uuid.UUID,
	cmd *command.SubscribePushCommand,
) (*common.PushSubscriptionResult, error) {
	if cmd == nil {
		return nil, xerrors.PushEndpointRequiredError()
	}
	if clientID == uuid.Nil() {
		return nil, xerrors.InvalidIDError()
	}

	validated, err := entity.NewValidated(entity.NewPushSubscription(entity.PushSubscriptionVariants{
		ClientID:  clientID,
		Endpoint:  cmd.Endpoint,
		P256dh:    cmd.P256dh,
		Auth:      cmd.Auth,
		UserAgent: cmd.UserAgent,
	}))
	if err != nil {
		return nil, fmt.Errorf("validate push subscription: %w", err)
	}

	saved, err := s.subscriptions.Upsert(ctx, validated)
	if err != nil {
		return nil, fmt.Errorf("upsert push subscription: %w", err)
	}

	return mapper.PushSubscriptionResult(saved), nil
}

func (s *service) Unsubscribe(ctx context.Context, clientID, subscriptionID uuid.UUID) error {
	if clientID == uuid.Nil() || subscriptionID == uuid.Nil() {
		return xerrors.InvalidIDError()
	}

	sub, err := s.subscriptions.GetByID(ctx, subscriptionID)
	if err != nil {
		if errors.Is(err, xerrors.ErrNotFound) {
			return xerrors.ErrNotFound
		}
		return fmt.Errorf("get push subscription: %w", err)
	}
	if sub.ClientID() != clientID {
		return xerrors.ErrNotFound
	}

	if err := s.subscriptions.DeleteByID(ctx, subscriptionID); err != nil {
		return fmt.Errorf("delete push subscription: %w", err)
	}
	return nil
}

func (s *service) Deliver(ctx context.Context, validatedCmmand entity.ValidatedPushCommand) error {
	command := validatedCmmand.Value

	clientID, err := s.keys.Authenticate(ctx, command.APIKey())
	if err != nil {
		if errors.Is(err, xerrors.ErrAuth) {
			return xerrors.DropMessageError()
		}
		return fmt.Errorf("authenticate push command: %w", err)
	}

	sub, err := s.subscriptions.GetByID(ctx, command.SubscriptionID())
	if err != nil {
		if errors.Is(err, xerrors.ErrNotFound) {
			return xerrors.DropMessageError()
		}
		return fmt.Errorf("get push subscription: %w", err)
	}
	if sub.ClientID() != clientID {
		return xerrors.DropMessageError()
	}

	if s.vapidPublic == "" || s.vapidPrivate == "" {
		return fmt.Errorf("send web push: %w", xerrors.VAPIDNotConfiguredError())
	}

	message := command.Message()
	result, err := s.notifier.Send(ctx, sub, &message)
	if err != nil {
		return fmt.Errorf("send web push: %w", err)
	}
	if result != nil && result.Expired {
		if err := s.subscriptions.DeleteByID(ctx, sub.ID()); err != nil && !errors.Is(err, xerrors.ErrNotFound) {
			return fmt.Errorf("delete expired subscription: %w", err)
		}
	}
	return nil
}

var _ interfaces.PushUseCase = (*service)(nil)
