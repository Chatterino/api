package twitchusernamecache

import (
	"context"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/nicklaw5/helix"
)

func New(ctx context.Context, cfg config.APIConfig, pool db.Pool, helixClient *helix.Client) cache.Cache {
	if helixClient == nil {
		return nil
	}

	usernameLoader := &UsernameLoader{
		helixClient: helixClient,
	}

	return cache.NewPostgreSQLCache(
		ctx, cfg, pool, cache.NewPrefixKeyProvider("twitch:username"), usernameLoader, 1*time.Hour,
	)
}
