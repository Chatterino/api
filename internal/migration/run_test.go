package migration

import (
	"context"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	qt "github.com/frankban/quicktest"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func TestRun(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()

	c.Run("No migrations", func(c *qt.C) {
		migrations = []Migration{}

		c.Run("No table created", func(c *qt.C) {
			pool.
				ExpectExec("CREATE TABLE IF NOT EXISTS").
				WithArgs().
				WillReturnResult(pgxmock.NewResult("CREATE TABLE", 1))
			pool.
				ExpectQuery("SELECT").
				WillReturnError(pgx.ErrNoRows)
			pool.
				ExpectExec("INSERT").
				WithArgs().
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
			oldVersion, newVersion, err := Run(ctx, pool)
			c.Assert(oldVersion, qt.Equals, int64(0))
			c.Assert(newVersion, qt.Equals, int64(0))
			c.Assert(err, qt.IsNil)
		})
		c.Run("Failed creating table", func(c *qt.C) {
			pool.
				ExpectExec("CREATE TABLE IF NOT EXISTS").
				WithArgs().
				WillReturnError(pgx.ErrNoRows)
			oldVersion, newVersion, err := Run(ctx, pool)
			c.Assert(oldVersion, qt.Equals, int64(0))
			c.Assert(newVersion, qt.Equals, int64(0))
			c.Assert(err, qt.ErrorIs, pgx.ErrNoRows)
		})
		c.Run("Failed getting version", func(c *qt.C) {
			pool.
				ExpectExec("CREATE TABLE IF NOT EXISTS").
				WithArgs().
				WillReturnResult(pgxmock.NewResult("CREATE TABLE", 1))
			pool.
				ExpectQuery("SELECT").
				WillReturnError(pgx.ErrTxClosed)
			oldVersion, newVersion, err := Run(ctx, pool)
			c.Assert(oldVersion, qt.Equals, int64(0))
			c.Assert(newVersion, qt.Equals, int64(0))
			c.Assert(err, qt.ErrorIs, pgx.ErrTxClosed)
		})
	})
	c.Run("1 migration", func(c *qt.C) {
		migrations = []Migration{
			{
				Version: 1,
				Up: func(ctx context.Context, tx pgx.Tx) error {
					return nil
				},
			},
		}
		pool.
			ExpectExec("CREATE TABLE IF NOT EXISTS").
			WithArgs().
			WillReturnResult(pgxmock.NewResult("CREATE TABLE", 1))
		pool.
			ExpectQuery("SELECT").
			WillReturnError(pgx.ErrNoRows)
		pool.
			ExpectExec("INSERT").
			WithArgs().
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		for _, m := range migrations {
			pool.
				ExpectBegin()
			pool.
				ExpectExec("UPDATE").
				WithArgs(m.Version).
				WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			pool.
				ExpectCommit()
		}
		oldVersion, newVersion, err := Run(ctx, pool)
		c.Assert(oldVersion, qt.Equals, int64(0))
		c.Assert(newVersion, qt.Equals, int64(1))
		c.Assert(err, qt.IsNil)
	})
	c.Run("2 migrations", func(c *qt.C) {
		migrations = []Migration{
			{
				Version: 1,
				Up: func(ctx context.Context, tx pgx.Tx) error {
					return nil
				},
			},
			{
				Version: 2,
				Up: func(ctx context.Context, tx pgx.Tx) error {
					return nil
				},
			},
		}
		pool.
			ExpectExec("CREATE TABLE IF NOT EXISTS").
			WithArgs().
			WillReturnResult(pgxmock.NewResult("CREATE TABLE", 1))
		pool.
			ExpectQuery("SELECT").
			WillReturnError(pgx.ErrNoRows)
		pool.
			ExpectExec("INSERT").
			WithArgs().
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		for _, m := range migrations {
			pool.
				ExpectBegin()
			pool.
				ExpectExec("UPDATE").
				WithArgs(m.Version).
				WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			pool.
				ExpectCommit()
		}
		oldVersion, newVersion, err := Run(ctx, pool)
		c.Assert(oldVersion, qt.Equals, int64(0))
		c.Assert(newVersion, qt.Equals, int64(2))
		c.Assert(err, qt.IsNil)
	})
	c.Run("Migration failed", func(c *qt.C) {
		migrations = []Migration{
			{
				Version: 1,
				Up: func(ctx context.Context, tx pgx.Tx) error {
					return pgx.ErrTxClosed
				},
			},
		}
		pool.
			ExpectExec("CREATE TABLE IF NOT EXISTS").
			WithArgs().
			WillReturnResult(pgxmock.NewResult("CREATE TABLE", 1))
		pool.
			ExpectQuery("SELECT").
			WillReturnError(pgx.ErrNoRows)
		pool.
			ExpectExec("INSERT").
			WithArgs().
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		for _, m := range migrations {
			pool.
				ExpectBegin()
			pool.
				ExpectExec("UPDATE").
				WithArgs(m.Version).
				WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			pool.
				ExpectCommit()
		}
		oldVersion, newVersion, err := Run(ctx, pool)
		c.Assert(oldVersion, qt.Equals, int64(0))
		c.Assert(newVersion, qt.Equals, int64(0))
		c.Assert(err, qt.ErrorIs, pgx.ErrTxClosed)
	})
}
