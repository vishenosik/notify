package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
	"github.com/vishenosik/notify/internal/domain/xerrors"
	"github.com/vishenosik/notify/internal/infrastructure/db/postgres/mapper"
	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
)

type apiKeyRepository struct {
	pool *Pool
}

func (p *Pool) APIKeyRepository() repository.APIKeyRepository {
	return &apiKeyRepository{
		pool: p,
	}
}

func (r *apiKeyRepository) q(ctx context.Context) db.Querier {
	return r.pool.queriesOrTx(ctx)
}

func (r *apiKeyRepository) Create(ctx context.Context, key entity.ValidatedAPIKey) (*entity.APIKey, error) {
	row, err := r.q(ctx).CreateAPIKey(ctx, mapper.ToDbCreateAPIKeyParams(key))
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", mapper.PgError(err))
	}

	return mapper.FromDbAPIKey(row), nil
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, hash []byte) (*entity.APIKey, error) {
	row, err := r.q(ctx).GetAPIKeyByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}

		return nil, fmt.Errorf("get api key by hash: %w", err)
	}

	return mapper.FromDbAPIKey(row), nil
}

func (r *apiKeyRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) (*entity.APIKey, error) {
	row, err := r.q(ctx).GetAPIKeyByClientId(ctx, clientID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}

		return nil, fmt.Errorf("get api key by client id: %w", err)
	}

	return mapper.FromDbAPIKey(row), nil
}

func (r *apiKeyRepository) ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*entity.APIKey, error) {
	rows, err := r.q(ctx).ListAPIKeysByClientId(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}

	return mapper.FromDbAPIKeys(rows), nil
}

func (r *apiKeyRepository) Delete(ctx context.Context, id uuid.UUID, archivedAt time.Time) error {
	n, err := r.q(ctx).DeleteAPIKey(
		ctx,
		mapper.ToDbDeleteAPIKeyParams(id, archivedAt),
	)
	if err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}

	if n == 0 {
		return xerrors.APIKeyNotFoundError()
	}

	return nil
}

func (r *apiKeyRepository) DeleteByClientID(
	ctx context.Context,
	clientID uuid.UUID,
	archivedAt time.Time,
) error {
	if _, err := r.q(ctx).DeleteAPIKeysByClientId(
		ctx,
		mapper.ToDbDeleteAPIKeysByClientIdParams(clientID, archivedAt),
	); err != nil {
		return fmt.Errorf("delete api keys by client: %w", err)
	}

	return nil
}
