package twitch

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

var clipSlugRegex = regexp.MustCompile(`^\/(\w{2,25}\/clip\/)?(clip\/)?([a-zA-Z0-9]+(?:-[-\w]{16})?)$`)

type ClipResolver struct {
	clipCache cache.Cache
}

func (r *ClipResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	// Regardless of domain path needs to match anyway, so we do it here to avoid duplication
	matches := clipSlugRegex.FindStringSubmatch(url.Path)

	match, domain := resolver.MatchesHosts(url, domains)
	if !match {
		return ctx, false
	}

	if len(matches) != 4 {
		return ctx, false
	}

	if domain == "m.twitch.tv" {
		// Find clips that look like https://m.twitch.tv/clip/SlugHere
		// matches[2] contains "clip/" - both this and matches[1] cannot be non-empty at the same time
		if matches[2] == "clip/" {
			return ctx, matches[1] == ""
		}

		// Find clips that look like https://m.twitch.tv/StreamerName/clip/SlugHere
		// matches[1] contains "StreamerName/clip/" - we need it in this check
		return ctx, matches[1] != ""
	}

	// Find clips that look like https://clips.twitch.tv/SlugHere
	if domain == "clips.twitch.tv" {
		// matches[1] contains "StreamerName/clip/" - we don't want it in this check though
		return ctx, matches[1] == ""
	}

	// Find clips that look like https://twitch.tv/StreamerName/clip/SlugHere
	// matches[1] contains "StreamerName/clip/" - we need it in this check
	return ctx, matches[1] != ""
}

func (r *ClipResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	clipSlug, err := parseClipSlug(url)
	if err != nil {
		return nil, err
	}

	return r.clipCache.Get(ctx, clipSlug, req)
}

func (r *ClipResolver) Name() string {
	return "twitch:clip"
}

func NewClipResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, helixAPI TwitchAPIClient) *ClipResolver {
	clipLoader := &ClipLoader{
		helixAPI: helixAPI,
	}

	r := &ClipResolver{
		clipCache: cache.NewPostgreSQLCache(
			ctx, cfg, pool, cache.NewPrefixKeyProvider("twitch:clip"),
			resolver.NewResponseMarshaller(clipLoader), 1*time.Hour,
		),
	}

	return r
}
