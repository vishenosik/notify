package postgres

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/repository"
	"github.com/vishenosik/notify/internal/domain/xerrors"
	"github.com/vishenosik/notify/internal/infrastructure/db/postgres/mapper"
	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
)

type pushSubscriptionRepository struct {
	pool *Pool
}

func (p *Pool) PushSubscriptionRepository() repository.PushSubscriptionRepository {
	return &pushSubscriptionRepository{
		pool: p,
	}
}

func (r *pushSubscriptionRepository) q(ctx context.Context) db.Querier {
	return r.pool.queriesOrTx(ctx)
}

func (r *pushSubscriptionRepository) Upsert(
	ctx context.Context,
	sub entity.ValidatedPushSubscription,
) (*entity.PushSubscription, error) {
	row, err := r.q(ctx).UpsertPushSubscription(ctx, mapper.ToDbUpsertPushSubscriptionParams(sub))
	if err != nil {
		return nil, fmt.Errorf("upsert push subscription: %w", mapper.PgError(err))
	}

	return mapper.FromDbPushSubscription(row), nil
}

func (r *pushSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PushSubscription, error) {
	row, err := r.q(ctx).GetPushSubscriptionById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}

		return nil, fmt.Errorf("get push subscription by id: %w", err)
	}

	return mapper.FromDbPushSubscription(row), nil
}

func (r *pushSubscriptionRepository) GetByClientIDAndEndpoint(
	ctx context.Context,
	clientID uuid.UUID,
	endpoint string,
) (*entity.PushSubscription, error) {
	row, err := r.q(ctx).GetPushSubscriptionByClientIdAndEndpoint(
		ctx,
		db.GetPushSubscriptionByClientIdAndEndpointParams{
			ClientID: clientID,
			Endpoint: endpoint,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}

		return nil, fmt.Errorf("get push subscription by endpoint: %w", err)
	}

	return mapper.FromDbPushSubscription(row), nil
}

func (r *pushSubscriptionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := r.q(ctx).DeletePushSubscriptionById(ctx, id); err != nil {
		return fmt.Errorf("delete push subscription: %w", err)
	}
	return nil
}

func (r *pushSubscriptionRepository) DeleteByClientIDAndEndpoint(
	ctx context.Context,
	clientID uuid.UUID,
	endpoint string,
) error {
	if err := r.q(ctx).DeletePushSubscriptionByClientIdAndEndpoint(
		ctx,
		mapper.ToDbDeletePushSubscriptionParams(clientID, endpoint),
	); err != nil {
		return fmt.Errorf("delete push subscription by endpoint: %w", err)
	}

	return nil
}

func (r *pushSubscriptionRepository) DeleteByClientID(ctx context.Context, clientID uuid.UUID) error {
	if err := r.q(ctx).DeletePushSubscriptionsByClientId(ctx, clientID); err != nil {
		return fmt.Errorf("delete push subscriptions by client: %w", err)
	}

	return nil
}
