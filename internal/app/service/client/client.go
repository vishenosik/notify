package client

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/app/mapper"
	"github.com/vishenosik/notify/internal/app/service/apikey"
	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

type service struct {
	clients       repository.ClientRepository
	keys          repository.APIKeyRepository
	subscriptions repository.PushSubscriptionRepository
	tx            repository.TxManager
}

func New(
	clients repository.ClientRepository,
	keys repository.APIKeyRepository,
	subscriptions repository.PushSubscriptionRepository,
	tx repository.TxManager,
) interfaces.ClientUseCase {
	return &service{
		clients:       clients,
		keys:          keys,
		subscriptions: subscriptions,
		tx:            tx,
	}
}

func (s *service) Create(ctx context.Context, cmd *command.CreateClientCommand) (*common.CreatedClientResult, error) {
	if cmd == nil {
		return nil, xerrors.ClientNameRequiredError()
	}

	validatedClient, err := entity.NewValidated(entity.NewClient(entity.ClientVariants{
		Name:        cmd.Name,
		Description: cmd.Description,
	}))
	if err != nil {
		return nil, fmt.Errorf("validate client: %w", err)
	}

	mat, err := apikey.GenerateMaterial()
	if err != nil {
		return nil, err
	}

	validatedKey, err := entity.NewValidated(entity.NewAPIKey(entity.APIKeyVariants{
		ClientID: validatedClient.Value.ID(),
		Name:     entity.APIKeyDefaultName,
		KeyHash:  mat.Hash(),
		Prefix:   mat.Prefix(),
		Last4:    mat.Last4(),
	}))
	if err != nil {
		return nil, fmt.Errorf("validate api key: %w", err)
	}

	var (
		savedClient *entity.Client
		savedKey    *entity.APIKey
	)
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		createdClient, err := s.clients.Create(ctx, validatedClient)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}
		createdKey, err := s.keys.Create(ctx, validatedKey)
		if err != nil {
			return fmt.Errorf("create api key: %w", err)
		}
		savedClient = createdClient
		savedKey = createdKey
		return nil
	}); err != nil {
		return nil, err
	}

	return mapper.CreatedClientResult(savedClient, savedKey, mat.Plaintext()), nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*common.ClientResult, error) {
	if id == uuid.Nil() {
		return nil, xerrors.InvalidIDError()
	}

	client, err := s.clients.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, xerrors.ErrNotFound) {
			return nil, xerrors.ClientNotFoundError()
		}
		return nil, fmt.Errorf("get client: %w", err)
	}

	return s.clientResult(ctx, client)
}

func (s *service) List(ctx context.Context) ([]*common.ClientResult, error) {
	clients, err := s.clients.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	out := make([]*common.ClientResult, 0, len(clients))
	for _, c := range clients {
		result, err := s.clientResult(ctx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, result)
	}
	return out, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil() {
		return xerrors.InvalidIDError()
	}

	archivedAt := time.Now().UTC()
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.clients.Delete(ctx, id, archivedAt); err != nil {
			return fmt.Errorf("archive client: %w", err)
		}
		if err := s.keys.DeleteByClientID(ctx, id, archivedAt); err != nil {
			return fmt.Errorf("archive api keys: %w", err)
		}
		if err := s.subscriptions.DeleteByClientID(ctx, id); err != nil {
			return fmt.Errorf("delete subscriptions: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *service) clientResult(ctx context.Context, client *entity.Client) (*common.ClientResult, error) {
	key, err := s.keys.GetByClientID(ctx, client.ID())
	if err != nil {
		if !errors.Is(err, xerrors.ErrNotFound) {
			return nil, fmt.Errorf("get api key: %w", err)
		}
		key = nil
	}
	return mapper.ClientResult(client, key), nil
}

var _ interfaces.ClientUseCase = (*service)(nil)
