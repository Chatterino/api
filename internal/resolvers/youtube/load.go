package youtube

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
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

	log.Println("[YouTube] GET", videoID)
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

func loadChannels(channelID string, r *http.Request) (interface{}, time.Duration, error) {
	youtubeChannelParts := []string{
		"statistics",
		"snippet",
	}

	log.Println("[YouTube] GET", channelID)
	builtRequest := youtubeClient.Channels.List(youtubeChannelParts)

	if strings.HasPrefix(channelID, "UC") {
		builtRequest = builtRequest.Id(channelID)
	} else {
		builtRequest = builtRequest.ForUsername(channelID)
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

	if channel.ContentDetails == nil {
		return &resolver.Response{Status: 500, Message: "channel unavailable"}, cache.NoSpecialDur, nil
	}

	data := youtubeChannelTooltipData{
		Title:        channel.Snippet.Title,
		PublishDate:  humanize.CreationDateRFC3339(channel.Snippet.PublishedAt),
		Views:        humanize.Number(channel.Statistics.ViewCount),
		Description:  channel.Snippet.Description,
		Subscribers:  humanize.Number(channel.Statistics.SubscriberCount),
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
