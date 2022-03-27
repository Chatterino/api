package oembed

import (
	"context"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/pashagolub/pgxmock"

	qt "github.com/frankban/quicktest"
)

func TestInitialize(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	pool, err := pgxmock.NewPool()
	c.Assert(err, qt.IsNil)

	c.Run("No file", func(c *qt.C) {
		dir := t.TempDir()
		cfg := config.APIConfig{
			OembedProvidersPath: dir + "/providers.json",
		}
		customResolvers := []resolver.Resolver{}
		c.Assert(customResolvers, qt.HasLen, 0)
		Initialize(ctx, cfg, pool, &customResolvers)
		c.Assert(customResolvers, qt.HasLen, 0)
	})
}
