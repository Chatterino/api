//go:build !test || migrationtest

package migration

import (
	"context"

	"github.com/jackc/pgx/v4"
)

func init() {
	// The version of this migration
	const migrationVersion = 1

	Register(
		migrationVersion,
		func(ctx context.Context, tx pgx.Tx) error {
			// The Up action of this migration
			_, err := tx.Exec(ctx, `
CREATE TABLE cache (
    key TEXT UNIQUE NOT NULL,
    value bytea NOT NULL,
    created_on TIMESTAMP NOT NULL DEFAULT now(),
    cached_until TIMESTAMP NOT NULL
);

CREATE INDEX idx_cache_key ON cache(key);
CREATE INDEX idx_cache_cached_until ON cache(cached_until);
		`)

			return err
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// The Down action of this migration
			_, err := tx.Exec(ctx, `
DROP TABLE cache;
		`)

			return err
		},
	)
}
