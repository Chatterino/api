package youtube

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

var youtubePlaylistRegex = regexp.MustCompile(`^/playlist$`)

type YouTubePlaylistResolver struct {
	playlistCache cache.Cache
}

func (r *YouTubePlaylistResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if !utils.IsSubdomainOf(url, "youtube.com") {
		return ctx, false
	}

	matches := youtubePlaylistRegex.MatchString(url.Path)
	return ctx, matches
}

func (r *YouTubePlaylistResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	log := logger.FromContext(ctx)

	q := url.Query()
	log.Infow("TEST", "q", q)

	playlistId := q.Get("list")
	if playlistId == "" {
		log.Warnw("[YouTube] Failed to get playlist ID from url",
			"url", url,
		)
		return nil, errors.New("no YouTube playlist id found")
	}

	return r.playlistCache.Get(ctx, fmt.Sprintf("playlist:%s", playlistId), req)
}

func (r *YouTubePlaylistResolver) Name() string {
	return "youtube:playlist"
}

func getPlaylistFromPath(path string) (string, error) {
	match := youtubePlaylistRegex.FindStringSubmatch(path)
	if match == nil || len(match) != 3 {
		return "", errors.New("invalid playlist")
	}

	return match[2], nil
}

func NewYouTubePlaylistResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, youtubeClient *youtubeAPI.Service) *YouTubePlaylistResolver {
	loader := NewYouTubePlaylistLoader(youtubeClient)

	r := &YouTubePlaylistResolver{
		playlistCache: cache.NewPostgreSQLCache(
			ctx, cfg, pool, cache.NewPrefixKeyProvider("youtube:playlist"), loader, cfg.YoutubeChannelCacheDuration,
		),
	}

	return r
}
