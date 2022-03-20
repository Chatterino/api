package imgur

import (
	"context"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	qt "github.com/frankban/quicktest"
	"github.com/pashagolub/pgxmock"
)

func TestInitialize(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	pool, err := pgxmock.NewPool()
	c.Assert(err, qt.IsNil)

	c.Run("No credentials", func(c *qt.C) {
		cfg := config.APIConfig{}
		customResolvers := []resolver.Resolver{}
		c.Assert(customResolvers, qt.HasLen, 0)
		Initialize(ctx, cfg, pool, &customResolvers)
		c.Assert(customResolvers, qt.HasLen, 0)
	})

	c.Run("Credentials", func(c *qt.C) {
		cfg := config.APIConfig{
			ImgurClientID: "test",
		}
		customResolvers := []resolver.Resolver{}
		c.Assert(customResolvers, qt.HasLen, 0)
		Initialize(ctx, cfg, pool, &customResolvers)
		c.Assert(customResolvers, qt.HasLen, 1)
	})
}
