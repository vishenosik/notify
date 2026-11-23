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

type clientRepository struct {
	pool *Pool
}

func (p *Pool) ClientRepository() repository.ClientRepository {
	return &clientRepository{
		pool: p,
	}
}

func (r *clientRepository) q(ctx context.Context) db.Querier {
	return r.pool.queriesOrTx(ctx)
}

func (r *clientRepository) Create(ctx context.Context, client entity.ValidatedClient) (*entity.Client, error) {
	row, err := r.q(ctx).CreateClient(ctx, mapper.ToDbCreateClientParams(client))
	if err != nil {
		return nil, fmt.Errorf("create client: %w", mapper.PgError(err))
	}

	return mapper.FromDbClient(row), nil
}

func (r *clientRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Client, error) {
	row, err := r.q(ctx).GetClientById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}

		return nil, fmt.Errorf("get client by id: %w", err)
	}

	return mapper.FromDbClient(row), nil
}

func (r *clientRepository) GetByName(ctx context.Context, name string) (*entity.Client, error) {
	row, err := r.q(ctx).GetClientByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}

		return nil, fmt.Errorf("get client by name: %w", err)
	}

	return mapper.FromDbClient(row), nil
}

func (r *clientRepository) List(ctx context.Context) ([]*entity.Client, error) {
	rows, err := r.q(ctx).ListClients(ctx)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	return mapper.FromDbClientList(rows), nil
}

func (r *clientRepository) Delete(ctx context.Context, id uuid.UUID, archivedAt time.Time) error {
	n, err := r.q(ctx).DeleteClient(ctx, mapper.ToDbDeleteClientParams(id, archivedAt))
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}

	if n == 0 {
		return xerrors.ClientNotFoundError()
	}

	return nil
}
