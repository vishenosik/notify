package apikey

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/vishenosik/notify/internal/app/common"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/app/mapper"
	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

type service struct {
	clients repository.ClientRepository
	keys    repository.APIKeyRepository
	tx      repository.TxManager
}

func New(
	clients repository.ClientRepository,
	keys repository.APIKeyRepository,
	tx repository.TxManager,
) interfaces.APIKeyUseCase {
	return &service{
		clients: clients,
		keys:    keys,
		tx:      tx,
	}
}

func (s *service) Authenticate(ctx context.Context, rawKey string) (uuid.UUID, error) {
	rawKey = strings.TrimSpace(rawKey)
	if !strings.HasPrefix(rawKey, keyPrefix) {
		return uuid.Nil(), xerrors.ErrAuth
	}

	key, err := s.keys.GetByHash(ctx, hashKey(rawKey))
	if err != nil {
		if errors.Is(err, xerrors.ErrNotFound) {
			return uuid.Nil(), xerrors.ErrAuth
		}
		return uuid.Nil(), fmt.Errorf("lookup api key: %w", err)
	}

	return key.ClientID(), nil
}

func (s *service) Rotate(ctx context.Context, clientID uuid.UUID) (*common.CreatedClientResult, error) {
	if clientID == uuid.Nil() {
		return nil, xerrors.InvalidIDError()
	}

	client, err := s.clients.GetByID(ctx, clientID)
	if err != nil {
		if errors.Is(err, xerrors.ErrNotFound) {
			return nil, xerrors.ClientNotFoundError()
		}
		return nil, fmt.Errorf("get client: %w", err)
	}

	mat, err := generateMaterial()
	if err != nil {
		return nil, err
	}

	validated, err := entity.NewValidated(entity.NewAPIKey(entity.APIKeyVariants{
		ClientID: client.ID(),
		Name:     entity.APIKeyDefaultName,
		KeyHash:  mat.hash,
		Prefix:   mat.prefix,
		Last4:    mat.last4,
	}))
	if err != nil {
		return nil, fmt.Errorf("validate api key: %w", err)
	}

	var saved *entity.APIKey
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.keys.DeleteByClientID(ctx, client.ID(), time.Now().UTC()); err != nil {
			return fmt.Errorf("archive api keys: %w", err)
		}
		created, err := s.keys.Create(ctx, validated)
		if err != nil {
			return fmt.Errorf("create api key: %w", err)
		}
		saved = created
		return nil
	}); err != nil {
		return nil, err
	}

	return mapper.CreatedClientResult(client, saved, mat.plaintext), nil
}

var _ interfaces.APIKeyUseCase = (*service)(nil)
