package supinic

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type TrackResolver struct {
	trackCache cache.Cache
}

func (r *TrackResolver) Check(ctx context.Context, url *url.URL) bool {
	fmt.Println("Checking supinic", url)
	if !utils.IsDomains(url, trackListDomains) {
		return false
	}

	if !trackPathRegex.MatchString(url.Path) {
		return false
	}

	return true
}

func (r *TrackResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	matches := trackPathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, errInvalidTrackPath
	}

	trackID := matches[1]

	return r.trackCache.Get(ctx, trackID, req)
}

func NewTrackResolver(ctx context.Context, cfg config.APIConfig) *TrackResolver {
	trackLoader := &TrackLoader{}

	r := &TrackResolver{
		trackCache: cache.NewPostgreSQLCache(ctx, cfg, "supinic:track", resolver.NewResponseMarshaller(trackLoader), 1*time.Hour),
	}

	return r
}
