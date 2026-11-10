package middleware

import (
	"context"
	"uuid"
)

type clientIDCtxKey struct{}

func WithClientID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, clientIDCtxKey{}, id)
}

func ClientIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(clientIDCtxKey{}).(uuid.UUID)
	if !ok || id == uuid.Nil() {
		return uuid.Nil(), false
	}
	return id, true
}
