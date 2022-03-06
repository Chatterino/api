package seventv

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

type EmoteResolver struct {
	emoteCache cache.Cache
}

func (r *EmoteResolver) Check(ctx context.Context, url *url.URL) bool {
	if match, _ := resolver.MatchesHosts(url, domains); !match {
		return false
	}

	return emotePathRegex.MatchString(url.Path)
}

func (r *EmoteResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
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

func NewEmoteResolver(ctx context.Context, cfg config.APIConfig) *EmoteResolver {
	emoteLoader := &EmoteLoader{
		baseURL: cfg.BaseURL,
	}

	r := &EmoteResolver{
		emoteCache: cache.NewPostgreSQLCache(ctx, cfg, "seventv:emotes", resolver.NewResponseMarshaller(emoteLoader), 1*time.Hour),
	}

	return r
}
