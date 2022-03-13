package youtube

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type YouTubeVideoResolver struct {
	videoCache cache.Cache
}

func (r *YouTubeVideoResolver) Check(ctx context.Context, url *url.URL) bool {
	return utils.IsSubdomainOf(url, "youtube.com")
}

func (r *YouTubeVideoResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	videoID := getYoutubeVideoIDFromURL(url)

	if videoID == "" {
		return resolver.NoLinkInfoFound, nil
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
