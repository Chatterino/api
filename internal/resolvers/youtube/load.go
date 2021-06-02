package youtube

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

func loadVideos(videoID string, r *http.Request) (interface{}, time.Duration, error) {
	youtubeVideoParts := []string{
		"statistics",
		"snippet",
		"contentDetails",
	}

	log.Println("[YouTube] GET video", videoID)
	youtubeResponse, err := youtubeClient.Videos.List(youtubeVideoParts).Id(videoID).Do()
	if err != nil {
		return &resolver.Response{
			Status:  500,
			Message: "youtube api error " + resolver.CleanResponse(err.Error()),
		}, 1 * time.Hour, nil
	}

	if len(youtubeResponse.Items) != 1 {
		return nil, cache.NoSpecialDur, errors.New("videos response is not size 1")
	}

	video := youtubeResponse.Items[0]

	if video.ContentDetails == nil {
		return &resolver.Response{Status: 500, Message: "video unavailable"}, cache.NoSpecialDur, nil
	}

	data := youtubeVideoTooltipData{
		Title:        video.Snippet.Title,
		ChannelTitle: video.Snippet.ChannelTitle,
		Duration:     humanize.DurationPT(video.ContentDetails.Duration),
		PublishDate:  humanize.CreationDateRFC3339(video.Snippet.PublishedAt),
		Views:        humanize.Number(video.Statistics.ViewCount),
		LikeCount:    humanize.Number(video.Statistics.LikeCount),
		DislikeCount: humanize.Number(video.Statistics.DislikeCount),
	}

	var tooltip bytes.Buffer
	if err := youtubeVideoTooltipTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "youtube template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	thumbnail := video.Snippet.Thumbnails.Default.Url
	if video.Snippet.Thumbnails.Medium != nil {
		thumbnail = video.Snippet.Thumbnails.Medium.Url
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnail,
	}, cache.NoSpecialDur, nil
}

func loadChannels(channelCacheKey string, r *http.Request) (interface{}, time.Duration, error) {
	youtubeChannelParts := []string{
		"statistics",
		"snippet",
	}

	log.Println("[YouTube] GET channel", channelCacheKey)
	builtRequest := youtubeClient.Channels.List(youtubeChannelParts)

	channelId := deconstructChannelIdFromCacheKey(channelCacheKey)
	switch channelId.channelType {
		case UserChannel:
			builtRequest = builtRequest.ForUsername(channelId.id)
		case IdentifierChannel:
			builtRequest = builtRequest.Id(channelId.id)
		case InvalidChannel:
			return &resolver.Response{
				Status:  500,
				Message: "cached channel ID is invalid",
			}, 1 * time.Hour, nil
	}

	youtubeResponse, err := builtRequest.Do()

	if err != nil {
		return &resolver.Response{
			Status:  500,
			Message: "youtube api error " + resolver.CleanResponse(err.Error()),
		}, 1 * time.Hour, nil
	}

	if len(youtubeResponse.Items) != 1 {
		return nil, cache.NoSpecialDur, errors.New("channel response is not size 1")
	}

	channel := youtubeResponse.Items[0]

	data := youtubeChannelTooltipData{
		Title:        channel.Snippet.Title,
		PublishDate:  humanize.CreationDateRFC3339(channel.Snippet.PublishedAt),
		Description:  channel.Snippet.Description,
		Subscribers:  humanize.Number(channel.Statistics.SubscriberCount),
		// TODO: fix billions showing as millions (e.g. 2B shows as 2000M)
		Views:        humanize.Number(channel.Statistics.ViewCount),
	}

	var tooltip bytes.Buffer
	if err := youtubeChannelTooltipTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "youtube template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	thumbnail := channel.Snippet.Thumbnails.Default.Url
	if channel.Snippet.Thumbnails.Medium != nil {
		thumbnail = channel.Snippet.Thumbnails.Medium.Url
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnail,
	}, cache.NoSpecialDur, nil
}
