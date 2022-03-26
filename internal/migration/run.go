package migration

import (
	"context"
	"fmt"

	"github.com/Chatterino/api/internal/db"
	"github.com/jackc/pgx/v4"
)

type MigrationFunction func(ctx context.Context, tx pgx.Tx) error

type Migration struct {
	Version int64
	Up      MigrationFunction
	Down    MigrationFunction
}

func (m *Migration) MigrateTo(ctx context.Context, pool db.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	if err := m.Up(ctx, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if _, err := tx.Exec(ctx, `UPDATE migrations SET version=$1`, m.Version); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func createMigrationsTableIfItDoesNotAlreadyExist(ctx context.Context, pool db.Pool) error {
	const createTableQuery = `CREATE TABLE IF NOT EXISTS migrations(version BIGINT);`
	_, err := pool.Exec(ctx, createTableQuery)
	return err
}

func insertVersionRow(ctx context.Context, pool db.Pool) error {
	const query = `INSERT INTO migrations (version) VALUES (0)`
	_, err := pool.Exec(ctx, query)
	return err
}

// getCurrentVersion returns the current version from the database. if no version row is there, it will insert a new row with the default value 0
func getCurrentVersion(ctx context.Context, pool db.Pool) (int64, error) {
	const query = `SELECT version FROM migrations;`
	row := pool.QueryRow(ctx, query)
	var currentVersion int64

	if err := row.Scan(&currentVersion); err != nil {
		if err == pgx.ErrNoRows {
			err = insertVersionRow(ctx, pool)
		}
		return 0, err
	}

	return currentVersion, nil
}

func Run(ctx context.Context, pool db.Pool) (int64, int64, error) {
	if err := createMigrationsTableIfItDoesNotAlreadyExist(ctx, pool); err != nil {
		return 0, 0, fmt.Errorf("error creating migrations table: %w", err)
	}

	oldVersion, err := getCurrentVersion(ctx, pool)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting current version: %w", err)
	}

	newVersion := oldVersion

	relevantMigrations := getMigrations(migrations, oldVersion)

	for _, migration := range relevantMigrations {
		if err := migration.MigrateTo(ctx, pool); err != nil {
			return 0, 0, err
		}

		newVersion = migration.Version
	}

	return oldVersion, newVersion, nil
}
