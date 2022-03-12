package frankerfacez

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
	if match, _ := resolver.MatchesHosts(url, domains); !match {
		return false
	}

	if !emotePathRegex.MatchString(url.Path) {
		return false
	}

	return true
}

func (r *EmoteResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	matches := emotePathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 4 {
		return nil, errInvalidFrankerFaceZEmotePath
	}

	emoteHash := matches[1]

	return r.emoteCache.Get(ctx, emoteHash, req)
}

func (r *EmoteResolver) Name() string {
	return "frankerfacez:emote"
}

func NewEmoteResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool) *EmoteResolver {
	const emoteAPIURL = "https://api.frankerfacez.com/v1/emote/%s"

	emoteLoader := &EmoteLoader{
		emoteAPIURL: emoteAPIURL,
	}

	r := &EmoteResolver{
		emoteCache: cache.NewPostgreSQLCache(ctx, cfg, pool, "frankerfacez:emote", resolver.NewResponseMarshaller(emoteLoader), 1*time.Hour),
	}

	return r
}
