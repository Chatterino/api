package twitch

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

var userRegex = regexp.MustCompile(`^\/([a-zA-Z0-9_]+)$`)

type UserResolver struct {
	userCache cache.Cache
}

func (r *UserResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if !utils.IsDomain(url, "twitch.tv") {
		return ctx, false
	}

	userMatch := userRegex.FindStringSubmatch(url.Path)
	if len(userMatch) != 2 {
		return ctx, false
	}

	return ctx, true
}

func (r *UserResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	return r.userCache.Get(ctx, strings.ToLower(strings.TrimLeft(url.Path, "/")), req)
}

func (r *UserResolver) Name() string {
	return "twitch:user"
}

func NewUserResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, helixAPI TwitchAPIClient) *UserResolver {
	userLoader := &UserLoader{helixAPI: helixAPI}

	r := &UserResolver{
		userCache: cache.NewPostgreSQLCache(ctx, cfg, pool, cache.NewPrefixKeyProvider("twitch:user"),
			resolver.NewResponseMarshaller(userLoader), cfg.TwitchUsernameCacheDuration),
	}

	return r
}
