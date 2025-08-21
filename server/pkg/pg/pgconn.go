package pg

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// NewConnection establishes a new database connection pool using the provided context,
// connection configuration, and logger. Returns a pgxpool.Pool instance on success,
// or nil if the connection fails.
func NewConnection(ctx context.Context, cnf *connConfig) (*pgxpool.Pool, error) {
	masterDsn := cnf.getDsn()

	pg, err := pgxpool.Connect(ctx, masterDsn)
	if err != nil {
		return nil, err
	}

	return pg, nil
}
