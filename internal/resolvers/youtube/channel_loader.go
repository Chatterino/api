package youtube

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

type youtubeChannelTooltipData struct {
	Title       string
	JoinedDate  string
	Subscribers string
	Views       string
}

type YouTubeChannelLoader struct {
	youtubeClient *youtubeAPI.Service
}

func (r *YouTubeChannelLoader) Load(ctx context.Context, channelCacheKey string, req *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)
	youtubeChannelParts := []string{
		"statistics",
		"snippet",
	}

	log.Debugw("[YouTube] GET channel",
		"cacheKey", channelCacheKey,
	)
	builtRequest := r.youtubeClient.Channels.List(youtubeChannelParts)

	channel := getChannelFromCacheKey(channelCacheKey)
	if channel.Type == CustomChannel {
		// Channels with custom URLs aren't searchable with the channel/list endpoint
		// The only average way to do this at the moment is to do a YouTube search of that name
		// and filter for channels. Not ideal...

		searchRequest := r.youtubeClient.Search.List([]string{"snippet"}).Q(channel.ID).Type("channel")
		response, err := searchRequest.MaxResults(1).Do()

		if err != nil {
			return resolver.Errorf("YouTube search API error: %s", err)
		}

		if len(response.Items) == 0 {
			return &resolver.Response{
				Status:  404,
				Message: fmt.Sprintf("No YouTube channel with the ID %s found", resolver.CleanResponse(channel.ID)),
			}, 24 * time.Hour, nil
		}

		if len(response.Items) > 1 {
			return resolver.Errorf("YouTube search response contained %d items", len(response.Items))
		}

		channel.ID = response.Items[0].Snippet.ChannelId
	}

	switch channel.Type {
	case UserChannel:
		builtRequest = builtRequest.ForUsername(channel.ID)
	case IdentifierChannel:
		builtRequest = builtRequest.Id(channel.ID)
	case CustomChannel:
		builtRequest = builtRequest.Id(channel.ID)
	case InvalidChannel:
		return resolver.Errorf("YouTube API channel type is invalid for key: %s", channelCacheKey)
	}

	youtubeResponse, err := builtRequest.Do()

	if err != nil {
		return resolver.Errorf("YouTube API error: %s", err)
	}

	if len(youtubeResponse.Items) == 0 {
		return &resolver.Response{
			Status:  404,
			Message: fmt.Sprintf("No YouTube channel with the ID %s found", resolver.CleanResponse(channel.ID)),
		}, 24 * time.Hour, nil
	}

	if len(youtubeResponse.Items) > 1 {
		return resolver.Errorf("YouTube channel response contained %d items", len(youtubeResponse.Items))
	}

	youtubeChannel := youtubeResponse.Items[0]

	data := youtubeChannelTooltipData{
		Title:       youtubeChannel.Snippet.Title,
		JoinedDate:  humanize.CreationDateRFC3339(youtubeChannel.Snippet.PublishedAt),
		Subscribers: humanize.Number(youtubeChannel.Statistics.SubscriberCount),
		Views:       humanize.Number(youtubeChannel.Statistics.ViewCount),
	}

	var tooltip bytes.Buffer
	if err := youtubeChannelTooltipTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "youtube template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	thumbnail := youtubeChannel.Snippet.Thumbnails.Default.Url
	if youtubeChannel.Snippet.Thumbnails.Medium != nil {
		thumbnail = youtubeChannel.Snippet.Thumbnails.Medium.Url
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnail,
	}, cache.NoSpecialDur, nil
}

func NewYouTubeChannelLoader(youtubeClient *youtubeAPI.Service) *YouTubeChannelLoader {
	loader := &YouTubeChannelLoader{
		youtubeClient: youtubeClient,
	}

	return loader
}
