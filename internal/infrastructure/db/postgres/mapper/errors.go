package mapper

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func PgError(err error) error {
	if err == nil {
		return nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return xerrors.ErrConflict
		}
	}

	return err
}
