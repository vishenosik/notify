package mapper

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func TestMapPgError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "unique violation maps to conflict",
			err:  &pgconn.PgError{Code: "23505"},
			want: xerrors.ErrConflict,
		},
		{
			name: "other pg error preserved",
			err:  &pgconn.PgError{Code: "23503"},
			want: &pgconn.PgError{Code: "23503"},
		},
		{
			name: "non pg error preserved",
			err:  errors.New("connection reset"),
			want: errors.New("connection reset"),
		},
		{
			name: "nil stays nil",
			err:  nil,
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := PgError(tc.err)
			if tc.want == nil {
				require.NoError(t, got)
				return
			}
			if errors.Is(got, tc.want) {
				return
			}
			require.EqualError(t, got, tc.want.Error())
		})
	}
}
