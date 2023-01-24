package seventv

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

type EmoteResolver struct {
	emoteCache cache.Cache
}

func (r *EmoteResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if match, _ := resolver.MatchesHosts(url, domains); !match {
		return ctx, false
	}

	return ctx, emotePathRegex.MatchString(url.Path)
}

func (r *EmoteResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	matches := emotePathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, errInvalidSevenTVEmotePath
	}

	emoteHash := matches[1]

	return r.emoteCache.Get(ctx, emoteHash, req)
}

func (r *EmoteResolver) Name() string {
	return "seventv:emote"
}

func NewEmoteResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, apiURL *url.URL) *EmoteResolver {
	emoteLoader := NewEmoteLoader(cfg, apiURL)

	r := &EmoteResolver{
		emoteCache: cache.NewPostgreSQLCache(
			ctx, cfg, pool, cache.NewPrefixKeyProvider("seventv:emote"),
			resolver.NewResponseMarshaller(emoteLoader), cfg.SeventvEmoteCacheDuration),
	}

	return r
}
