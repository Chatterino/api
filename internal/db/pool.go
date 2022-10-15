//go:generate mockgen -destination ../mocks/mock_dbPool.go -package=mocks . Pool

package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Pool interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Begin(context.Context) (pgx.Tx, error)
}

func NewPool(ctx context.Context, dsn string) (Pool, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging to pool: %w", err)
	}

	return pool, nil
}
