package youtube

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

type YouTubeVideoShortURLResolver struct {
	videoCache cache.Cache
}

func (r *YouTubeVideoShortURLResolver) Check(ctx context.Context, url *url.URL) bool {
	return url.Host == "youtu.be"
}

func (r *YouTubeVideoShortURLResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	videoID := getYoutubeVideoIDFromURL2(url)

	if videoID == "" {
		return resolver.NoLinkInfoFound, nil
	}

	return r.videoCache.Get(ctx, videoID, req)
}

func NewYouTubeVideoShortURLResolver(videoCache cache.Cache) *YouTubeVideoShortURLResolver {
	r := &YouTubeVideoShortURLResolver{
		videoCache: videoCache,
	}

	return r
}
