package discord

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

type InviteResolver struct {
	inviteCache cache.Cache
}

func (r *InviteResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	return ctx, discordInviteURLRegex.MatchString(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
}

func (r *InviteResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	matches := discordInviteURLRegex.FindStringSubmatch(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
	if len(matches) != 4 {
		return nil, errInvalidDiscordInvite
	}

	inviteCode := matches[3]

	return r.inviteCache.Get(ctx, inviteCode, req)
}

func (r *InviteResolver) Name() string {
	return "discord:invite"
}

func NewInviteResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, baseURL *url.URL) *InviteResolver {
	inviteLoader := NewInviteLoader(baseURL, cfg.DiscordToken)

	// We cache invites longer on purpose as the API is pretty strict with its rate limiting, and the information changes very seldomly anyway
	// TODO: Log 429 errors from the loader
	r := &InviteResolver{
		inviteCache: cache.NewPostgreSQLCache(
			ctx, cfg, pool, cache.NewPrefixKeyProvider("discord:invite"),
			resolver.NewResponseMarshaller(inviteLoader), cfg.DiscordInviteCacheDuration),
	}

	return r
}
