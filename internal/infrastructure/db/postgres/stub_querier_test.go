package postgres

import (
	"context"
	"errors"
	"uuid"

	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
)

var errUnimplemented = errors.New("unimplemented")

type stubQuerier struct {
	createClient  func(context.Context, db.CreateClientParams) (db.Client, error)
	getClientByID func(context.Context, uuid.UUID) (db.Client, error)
}

func (s *stubQuerier) CreateAPIKey(context.Context, db.CreateAPIKeyParams) (db.ApiKey, error) {
	return db.ApiKey{}, errUnimplemented
}

func (s *stubQuerier) CreateClient(ctx context.Context, arg db.CreateClientParams) (db.Client, error) {
	if s.createClient != nil {
		return s.createClient(ctx, arg)
	}
	return db.Client{}, errUnimplemented
}

func (s *stubQuerier) DeleteAPIKey(context.Context, db.DeleteAPIKeyParams) (int64, error) {
	return 0, errUnimplemented
}

func (s *stubQuerier) DeleteAPIKeysByClientId(context.Context, db.DeleteAPIKeysByClientIdParams) (int64, error) {
	return 0, errUnimplemented
}

func (s *stubQuerier) DeleteClient(context.Context, db.DeleteClientParams) (int64, error) {
	return 0, errUnimplemented
}

func (s *stubQuerier) DeletePushSubscriptionByClientIdAndEndpoint(context.Context, db.DeletePushSubscriptionByClientIdAndEndpointParams) error {
	return errUnimplemented
}

func (s *stubQuerier) DeletePushSubscriptionById(context.Context, uuid.UUID) error {
	return errUnimplemented
}

func (s *stubQuerier) DeletePushSubscriptionsByClientId(context.Context, uuid.UUID) error {
	return errUnimplemented
}

func (s *stubQuerier) GetAPIKeyByClientId(context.Context, uuid.UUID) (db.ApiKey, error) {
	return db.ApiKey{}, errUnimplemented
}

func (s *stubQuerier) GetAPIKeyByHash(context.Context, []byte) (db.ApiKey, error) {
	return db.ApiKey{}, errUnimplemented
}

func (s *stubQuerier) GetClientById(ctx context.Context, id uuid.UUID) (db.Client, error) {
	if s.getClientByID != nil {
		return s.getClientByID(ctx, id)
	}
	return db.Client{}, errUnimplemented
}

func (s *stubQuerier) GetClientByName(context.Context, string) (db.Client, error) {
	return db.Client{}, errUnimplemented
}

func (s *stubQuerier) GetPushSubscriptionByClientIdAndEndpoint(context.Context, db.GetPushSubscriptionByClientIdAndEndpointParams) (db.PushSubscription, error) {
	return db.PushSubscription{}, errUnimplemented
}

func (s *stubQuerier) GetPushSubscriptionById(context.Context, uuid.UUID) (db.PushSubscription, error) {
	return db.PushSubscription{}, errUnimplemented
}

func (s *stubQuerier) ListAPIKeysByClientId(context.Context, uuid.UUID) ([]db.ApiKey, error) {
	return nil, errUnimplemented
}

func (s *stubQuerier) ListClients(context.Context) ([]db.Client, error) {
	return nil, errUnimplemented
}

func (s *stubQuerier) ListPushSubscriptionsByClientId(context.Context, uuid.UUID) ([]db.PushSubscription, error) {
	return nil, errUnimplemented
}

func (s *stubQuerier) UpsertPushSubscription(context.Context, db.UpsertPushSubscriptionParams) (db.PushSubscription, error) {
	return db.PushSubscription{}, errUnimplemented
}

var _ db.Querier = (*stubQuerier)(nil)
