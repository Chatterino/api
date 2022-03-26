//go:build !test || migrationtest

package migration

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestRealMigrations(t *testing.T) {
	c := qt.New(t)

	c.Run("Ensure migrations have been registered", func(c *qt.C) {
		c.Assert(migrations, qt.Not(qt.HasLen), 0)
	})

	c.Run("Ensure migrations don't reuse the version number", func(c *qt.C) {
		hitVersions := map[int64]struct{}{}
		for _, m := range migrations {
			if _, ok := hitVersions[m.Version]; ok {
				c.Fatalf("Migration %d was hit more than once", m.Version)
			}
			hitVersions[m.Version] = struct{}{}
		}
	})
}
