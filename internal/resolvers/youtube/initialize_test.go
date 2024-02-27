package youtube

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

	// google.golang.org/api/youtube/v3 automatically
	// uses this environment variable to initialize clients
	c.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")

	c.Run("No YouTube API key", func(c *qt.C) {
		cfg := config.APIConfig{
			YoutubeApiKey: "",
		}
		customResolvers := []resolver.Resolver{}
		c.Assert(customResolvers, qt.HasLen, 0)
		Initialize(ctx, cfg, pool, &customResolvers)
		c.Assert(customResolvers, qt.HasLen, 0)
	})
	c.Run("With YouTube API key", func(c *qt.C) {
		cfg := config.APIConfig{
			YoutubeApiKey: "test",
		}
		customResolvers := []resolver.Resolver{}
		c.Assert(customResolvers, qt.HasLen, 0)
		Initialize(ctx, cfg, pool, &customResolvers)
		c.Assert(customResolvers, qt.HasLen, 4)
	})
}
