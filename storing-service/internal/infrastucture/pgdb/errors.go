package pgdb

import (
	"errors"
	"storing-service/internal/errdefs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func handleDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errdefs.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return errdefs.ErrAlreadyExists
		case "23503", "23502", "23514":
			return errdefs.ErrInvalidArgument
		}
	}
	return err
}
