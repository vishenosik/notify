package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-retry"

	"github.com/vishenosik/notify/internal/domain/repository"
	"github.com/vishenosik/notify/internal/infrastructure/db/postgres/mapper"
	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
	"github.com/vishenosik/notify/pkg/logs"
)

const (
	poolMaxConns        = 8
	poolMinConns        = 0
	poolMaxConnLifetime = time.Hour
)

type Pool struct {
	pool    *pgxpool.Pool
	queries db.Querier
	logger  *slog.Logger
}

func NewPool(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	cfg.MaxConns = poolMaxConns
	cfg.MinConns = poolMinConns
	cfg.MaxConnLifetime = poolMaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	p := &Pool{
		pool:    pool,
		queries: db.New(pool),
		logger:  slog.Default().With(slog.String("component", "postgres-pool")),
	}

	go func() {
		if err := p.Ping(ctx); err != nil {
			p.logger.Error("ping database failed after retries", logs.Error(err))
			return
		}

		p.logger.Info("database connected")
	}()

	return p, nil
}

func (p *Pool) Close() error {
	p.pool.Close()
	return nil
}

//
// Pinger implementation
//

var _ repository.Pinger = (*Pool)(nil)

func (p *Pool) Name() string {
	return "database"
}

func (p *Pool) Ping(ctx context.Context) error {
	backoff := retry.NewFibonacci(1 * time.Second)
	backoff = retry.WithMaxRetries(10, backoff)

	if err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		if err := p.pool.Ping(ctx); err != nil {
			p.logger.Debug("ping database error", slog.Any("error", err))

			if _, ok := errors.AsType[*pgconn.ConnectError](err); ok {
				return retry.RetryableError(err)
			}

			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

//
// TxManager implementation
//

var _ repository.TxManager = (*Pool)(nil)

func (p *Pool) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			_ = rollbackErr
		}
	}()

	if err := fn(WithQueries(ctx, db.New(tx))); err != nil {
		return mapper.PgError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

type queriesCtxKey struct{}

func WithQueries(ctx context.Context, q db.Querier) context.Context {
	return context.WithValue(ctx, queriesCtxKey{}, q)
}

func QueriesFromContext(ctx context.Context) db.Querier {
	q, _ := ctx.Value(queriesCtxKey{}).(db.Querier)
	return q
}

func (p *Pool) queriesOrTx(ctx context.Context) db.Querier {
	if q := QueriesFromContext(ctx); q != nil {
		return q
	}
	return p.queries
}
