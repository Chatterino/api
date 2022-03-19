package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/utils"
)

type YouTubeVideoResolver struct {
	videoCache cache.Cache
}

func (r *YouTubeVideoResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if !utils.IsSubdomainOf(url, "youtube.com") {
		return ctx, false
	}

	return ctx, getYoutubeVideoIDFromURL(url) != ""
}

var (
	errInvalidVideoLink = errors.New("invalid video link")
)

func (r *YouTubeVideoResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	videoID := getYoutubeVideoIDFromURL(url)

	if videoID == "" {
		return nil, errInvalidVideoLink
	}

	return r.videoCache.Get(ctx, videoID, req)
}

func (r *YouTubeVideoResolver) Name() string {
	return "youtube:video"
}

func NewYouTubeVideoResolver(videoCache cache.Cache) *YouTubeVideoResolver {
	r := &YouTubeVideoResolver{
		videoCache: videoCache,
	}

	return r
}
