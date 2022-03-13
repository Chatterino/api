package livestreamfails

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type ClipResolver struct {
	clipCache cache.Cache
}

func (r *ClipResolver) Check(ctx context.Context, url *url.URL) bool {
	if !utils.IsSubdomainOf(url, "livestreamfails.com") {
		return false
	}

	if !pathRegex.MatchString(url.Path) {
		return false
	}

	return true
}

func (r *ClipResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	pathParts := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
	clipId := pathParts[1]

	return r.clipCache.Get(ctx, clipId, req)
}

func (r *ClipResolver) Name() string {
	return "livestreamfails:clip"
}

func NewClipResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool) *ClipResolver {
	clipLoader := &ClipLoader{}

	r := &ClipResolver{
		clipCache: cache.NewPostgreSQLCache(ctx, cfg, pool, "livestreamfails:clip", resolver.NewResponseMarshaller(clipLoader), 1*time.Hour),
	}

	return r
}
