//go:build !migrationtest

package migration

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/jackc/pgx/v4"
)

func TestRegister(t *testing.T) {
	emptyFunc := func(ctx context.Context, tx pgx.Tx) error {
		return nil
	}

	c := qt.New(t)

	c.Assert(migrations, qt.HasLen, 0)

	Register(1, emptyFunc, emptyFunc)

	c.Assert(migrations, qt.HasLen, 1)
}
