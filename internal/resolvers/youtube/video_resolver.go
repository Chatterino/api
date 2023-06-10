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

type contextKey string

var (
	contextVideoID = contextKey("videoID")

	errMissingVideoIDValue = errors.New("missing video ID in context")
)

func videoIDFromContext(ctx context.Context) (string, error) {
	videoID, ok := ctx.Value(contextVideoID).(string)
	if !ok {
		return "", errMissingVideoIDValue
	}

	return videoID, nil
}

func (r *YouTubeVideoResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if !utils.IsSubdomainOf(url, "youtube.com") {
		return ctx, false
	}

	videoID := getYoutubeVideoIDFromURL(url)

	ctx = context.WithValue(ctx, contextVideoID, videoID)

	return ctx, videoID != ""
}

var (
	errInvalidVideoLink = errors.New("invalid video link")
)

func (r *YouTubeVideoResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
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
