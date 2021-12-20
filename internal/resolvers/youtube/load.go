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

	// Check if a video is age resricted: https://stackoverflow.com/a/33750307
	var ageRestricted = false
	if video.ContentDetails.ContentRating != nil {
		if video.ContentDetails.ContentRating.YtRating == "ytAgeRestricted" {
			ageRestricted = true
		}
	}

	data := youtubeVideoTooltipData{
		Title:         video.Snippet.Title,
		ChannelTitle:  video.Snippet.ChannelTitle,
		Duration:      humanize.DurationPT(video.ContentDetails.Duration),
		PublishDate:   humanize.CreationDateRFC3339(video.Snippet.PublishedAt),
		Views:         humanize.Number(video.Statistics.ViewCount),
		LikeCount:     humanize.Number(video.Statistics.LikeCount),
		CommentCount:  humanize.Number(video.Statistics.CommentCount),
		AgeRestricted: ageRestricted,
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

	channelID := deconstructChannelIDFromCacheKey(channelCacheKey)
	if channelID.chanType == CustomChannel {
		// Channels with custom URLs aren't searchable with the channel/list endpoint
		// The only average way to do this at the moment is to do a YouTube search of that name
		// and filter for channels. Not ideal...

		searchRequest := youtubeClient.Search.List([]string{"snippet"}).Q(channelID.ID).Type("channel")
		response, err := searchRequest.MaxResults(1).Do()

		if err != nil {
			return &resolver.Response{
				Status:  500,
				Message: "youtube search api error " + resolver.CleanResponse(err.Error()),
			}, 1 * time.Hour, nil
		}

		if len(response.Items) != 1 {
			return nil, cache.NoSpecialDur, errors.New("channel search response is not size 1")
		}

		channelID.ID = response.Items[0].Snippet.ChannelId
	}

	switch channelID.chanType {
	case UserChannel:
		builtRequest = builtRequest.ForUsername(channelID.ID)
	case IdentifierChannel:
		builtRequest = builtRequest.Id(channelID.ID)
	case CustomChannel:
		builtRequest = builtRequest.Id(channelID.ID)
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
		Title:       channel.Snippet.Title,
		JoinedDate:  humanize.CreationDateRFC3339(channel.Snippet.PublishedAt),
		Subscribers: humanize.Number(channel.Statistics.SubscriberCount),
		Views:       humanize.Number(channel.Statistics.ViewCount),
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
