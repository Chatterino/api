package twitch

import (
	"context"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/mocks"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	qt "github.com/frankban/quicktest"
	"github.com/golang/mock/gomock"
	"github.com/pashagolub/pgxmock"
)

func TestInitialize(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)
	ctrl := gomock.NewController(c)
	helixClient := mocks.NewMockTwitchAPIClient(ctrl)
	defer ctrl.Finish()

	cfg := config.APIConfig{}
	pool, err := pgxmock.NewPool()
	c.Assert(err, qt.IsNil)

	c.Run("No helix client", func(c *qt.C) {
		customResolvers := []resolver.Resolver{}
		c.Assert(customResolvers, qt.HasLen, 0)
		Initialize(ctx, cfg, pool, nil, &customResolvers)
		c.Assert(customResolvers, qt.HasLen, 0)
	})
	c.Run("Helix client", func(c *qt.C) {
		customResolvers := []resolver.Resolver{}
		c.Assert(customResolvers, qt.HasLen, 0)
		Initialize(ctx, cfg, pool, helixClient, &customResolvers)
		c.Assert(customResolvers, qt.HasLen, 1)
	})
}
