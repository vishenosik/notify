package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
)

func TestQueriesFromContext(t *testing.T) {
	t.Parallel()

	t.Run("missing key returns nil", func(t *testing.T) {
		t.Parallel()
		require.Nil(t, QueriesFromContext(context.Background()))
	})

	t.Run("round-trip", func(t *testing.T) {
		t.Parallel()
		q := &db.Queries{}
		got := QueriesFromContext(WithQueries(context.Background(), q))
		require.Same(t, q, got)
	})
}
