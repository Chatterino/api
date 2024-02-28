package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/staticresponse"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

type youtubePlaylistTooltipData struct {
	Title       string
	Description string
	Channel     string
	VideoCount  int64
	PublishedAt string
}

type YouTubePlaylistLoader struct {
	youtubeClient *youtubeAPI.Service
}

func (r *YouTubePlaylistLoader) Load(ctx context.Context, playlistCacheKey string, req *http.Request) ([]byte, *int, *string, time.Duration, error) {
	log := logger.FromContext(ctx)
	log.Debugw("[YouTube] GET playlist",
		"cacheKey", playlistCacheKey,
	)

	playlistId, err := getPlaylistFromCacheKey(playlistCacheKey)
	if err != nil {
		return resolver.InternalServerErrorf("YouTube API playlist is invalid for key: %s", playlistCacheKey)
	}

	youtubePlaylistParts := []string{
		"snippet",
		"contentDetails",
	}

	youtubeResponse, err := r.youtubeClient.Playlists.List(youtubePlaylistParts).Id(playlistId).Do()
	if err != nil {
		return resolver.InternalServerErrorf("YouTube API error: %s", err)
	}

	if len(youtubeResponse.Items) == 0 {
		return staticresponse.NotFoundf("No YouTube playlist with the ID %s found", playlistId).
			WithCacheDuration(24 * time.Hour).
			Return()
	}

	if len(youtubeResponse.Items) > 1 {
		return resolver.InternalServerErrorf("YouTube playlist response contained %d items", len(youtubeResponse.Items))
	}

	youtubePlaylist := youtubeResponse.Items[0]

	data := youtubePlaylistTooltipData{
		Title:       youtubePlaylist.Snippet.Title,
		Description: youtubePlaylist.Snippet.Description,
		Channel:     youtubePlaylist.Snippet.ChannelTitle,
		VideoCount:  youtubePlaylist.ContentDetails.ItemCount,
		PublishedAt: youtubePlaylist.Snippet.PublishedAt,
	}

	var tooltip bytes.Buffer
	if err := youtubePlaylistTooltipTemplate.Execute(&tooltip, data); err != nil {
		return resolver.InternalServerErrorf("YouTube template error: %s", err.Error())
	}

	statusCode := http.StatusOK
	contentType := "application/json"

	response := &resolver.Response{
		Status:    statusCode,
		Tooltip:   tooltip.String(),
		Thumbnail: youtubePlaylist.Snippet.Thumbnails.Maxres.Url,
	}

	payload, err := json.Marshal(response)
	if err != nil {
		return resolver.InternalServerErrorf("YouTube marshaling error: %s", err.Error())
	}

	return payload, &statusCode, &contentType, cache.NoSpecialDur, nil
}

func getPlaylistFromCacheKey(cacheKey string) (string, error) {
	splitKey := strings.Split(cacheKey, ":")

	if len(splitKey) < 2 {
		return "", errors.New("invalid playlist")
	}

	return splitKey[1], nil
}

func NewYouTubePlaylistLoader(youtubeClient *youtubeAPI.Service) *YouTubePlaylistLoader {
	loader := &YouTubePlaylistLoader{
		youtubeClient: youtubeClient,
	}

	return loader
}
