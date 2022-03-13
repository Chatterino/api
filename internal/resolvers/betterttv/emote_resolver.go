package betterttv

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

type EmoteResolver struct {
	emoteCache cache.Cache
}

func (r *EmoteResolver) Check(ctx context.Context, url *url.URL) bool {
	// Ensure that the domain is either betterttv.com or www.betterttv as defined in the domains map in initialize.go
	if match, _ := resolver.MatchesHosts(url, domains); !match {
		return false
	}

	// Ensure that the path of the url matches the emote path regex as defined in initialize.go
	if !emotePathRegex.MatchString(url.Path) {
		return false
	}

	return true
}

func (r *EmoteResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	matches := emotePathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, ErrInvalidBTTVEmotePath
	}

	emoteHash := matches[1]

	return r.emoteCache.Get(ctx, emoteHash, req)
}

func (r *EmoteResolver) Name() string {
	return "betterttv:emote"
}

func NewEmoteResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, emoteAPIURL *url.URL) *EmoteResolver {
	emoteLoader := NewEmoteLoader(emoteAPIURL)

	r := &EmoteResolver{
		emoteCache: cache.NewPostgreSQLCache(ctx, cfg, pool, "betterttv:emotes", resolver.NewResponseMarshaller(emoteLoader), 1*time.Hour),
	}

	return r
}
