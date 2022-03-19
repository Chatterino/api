package youtube

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

var (
	youtubeChannelRegex = regexp.MustCompile(`^/(c\/|channel\/|user\/)?([a-zA-Z0-9\-]{1,})$`)
)

type YouTubeChannelResolver struct {
	channelCache cache.Cache
}

func (r *YouTubeChannelResolver) Check(ctx context.Context, url *url.URL) bool {
	if !utils.IsSubdomainOf(url, "youtube.com") {
		return false
	}

	q := url.Query()
	if q.Has("v") {
		return false
	}

	matches := youtubeChannelRegex.MatchString(url.Path)
	return matches
}

func (r *YouTubeChannelResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	log := logger.FromContext(ctx)
	channel := getChannelFromPath(url.Path)

	if channel.Type == InvalidChannel {
		log.Warnw("[YouTube] URL was incorrectly treated as a channel",
			"url", url,
		)
		return resolver.NoLinkInfoFound, nil
	}

	return r.channelCache.Get(ctx, channel.ToCacheKey(), req)
}

func (r *YouTubeChannelResolver) Name() string {
	return "youtube:channel"
}

func NewYouTubeChannelResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, youtubeClient *youtubeAPI.Service) *YouTubeChannelResolver {
	loader := NewYouTubeChannelLoader(youtubeClient)

	r := &YouTubeChannelResolver{
		channelCache: cache.NewPostgreSQLCache(ctx, cfg, pool, "youtube:channel", resolver.NewResponseMarshaller(loader), 24*time.Hour),
	}

	return r
}
