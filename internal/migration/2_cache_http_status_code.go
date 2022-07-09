//go:build !test || migrationtest

package migration

import (
	"context"

	"github.com/jackc/pgx/v4"
)

func init() {
	// The version of this migration
	const migrationVersion = 2

	Register(
		migrationVersion,
		func(ctx context.Context, tx pgx.Tx) error {
			// The Up action of this migration
			// Delete all cached entries
			_, err := tx.Exec(ctx, `TRUNCATE cache;`)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `
ALTER TABLE cache
	ADD http_status_code SMALLINT NOT NULL,
	ADD http_content_type TEXT NOT NULL
;`)

			return err
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// The Down action of this migration
			_, err := tx.Exec(ctx, `
ALTER TABLE cache
	DROP http_status_code,
	DROP http_content_type
;`)

			return err
		},
	)
}
