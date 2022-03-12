package youtube

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"time"

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
	matches := youtubeChannelRegex.MatchString(url.Path)
	return utils.IsSubdomainOf(url, "youtube.com") && matches
}

func (r *YouTubeChannelResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	channel := getChannelFromPath(url.Path)

	if channel.Type == InvalidChannel {
		return resolver.NoLinkInfoFound, nil
	}

	return r.channelCache.Get(ctx, channel.ToCacheKey(), req)
}

func (r *YouTubeChannelResolver) Name() string {
	return "youtube:channel"
}

func NewYouTubeChannelResolver(ctx context.Context, cfg config.APIConfig, youtubeClient *youtubeAPI.Service) *YouTubeChannelResolver {
	loader := NewYouTubeChannelLoader(youtubeClient)

	r := &YouTubeChannelResolver{
		channelCache: cache.NewPostgreSQLCache(ctx, cfg, "youtube_channels", resolver.NewResponseMarshaller(loader), 24*time.Hour),
	}

	return r
}
