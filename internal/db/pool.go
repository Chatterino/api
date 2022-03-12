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
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
}

func NewPool(ctx context.Context, dsn string) (Pool, error) {

	var err error

	pool, err := pgxpool.Connect(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("error connecting to pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging to pool: %w", err)
	}

	// conn, err := pool.Acquire(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("error acquiring connection from pool: %w", err)
	// }
	// defer conn.Release()

	// if oldVersion, newVersion, err := migration.Run(ctx, conn.Conn()); err != nil {
	// 	return nil, fmt.Errorf("error running database migrations: %w", err)
	// 	log.Fatalw("Error running database migrations",
	// 		"dsn", dsn,
	// 		"error", err,
	// 	)
	// } else {
	// 	if newVersion != oldVersion {
	// 		log.Infow("Ran database migrations",
	// 			"oldVersion", oldVersion,
	// 			"newVersion", newVersion,
	// 		)
	// 	}
	// }

	return pool, nil
}
