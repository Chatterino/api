package supinic

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type TrackResolver struct {
	trackCache cache.Cache
}

func (r *TrackResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if !utils.IsDomains(url, trackListDomains) {
		return ctx, false
	}

	if !trackPathRegex.MatchString(url.Path) {
		return ctx, false
	}

	return ctx, true
}

func (r *TrackResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	matches := trackPathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, errInvalidTrackPath
	}

	trackID := matches[1]

	return r.trackCache.Get(ctx, trackID, req)
}

func (r *TrackResolver) Name() string {
	return "supinic:track"
}

func NewTrackResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool) *TrackResolver {
	trackLoader := &TrackLoader{}

	r := &TrackResolver{
		trackCache: cache.NewPostgreSQLCache(ctx, cfg, pool, "supinic:track", resolver.NewResponseMarshaller(trackLoader), 1*time.Hour),
	}

	return r
}
