//go:build !test || migrationtest

package migration

import (
	"context"

	"github.com/jackc/pgx/v4"
)

func init() {
	// The version of this migration
	const migrationVersion = 3

	Register(
		migrationVersion,
		func(ctx context.Context, tx pgx.Tx) error {
			// The Up action of this migration
			_, err := tx.Exec(ctx, `
CREATE TABLE dependent_values (
    key TEXT UNIQUE NOT NULL,
    parent_key TEXT NOT NULL,
    value bytea NOT NULL,
    http_content_type TEXT NOT NULL,
    committed BOOLEAN NOT NULL DEFAULT FALSE,
    expiration_timestamp TIMESTAMP NOT NULL
);

CREATE INDEX idx_dependent_values_key ON dependent_values(key);
CREATE INDEX idx_dependent_values_parent_entry_key ON dependent_values(parent_key);
CREATE INDEX idx_dependent_values_committed ON dependent_values(committed);
		`)

			return err
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// The Down action of this migration
			_, err := tx.Exec(ctx, `
DROP TABLE dependent_values;
		`)

			return err
		},
	)
}
