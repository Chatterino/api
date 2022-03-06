package youtube

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

type VideoLoader struct {
	youtubeClient *youtubeAPI.Service
}

func (r *VideoLoader) Load(ctx context.Context, videoID string, req *http.Request) (*resolver.Response, time.Duration, error) {
	youtubeVideoParts := []string{
		"statistics",
		"snippet",
		"contentDetails",
	}

	log.Println("[YouTube] GET video", videoID)
	youtubeResponse, err := r.youtubeClient.Videos.List(youtubeVideoParts).Id(videoID).Do()
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

func NewVideoLoader(youtubeClient *youtubeAPI.Service) *VideoLoader {
	loader := &VideoLoader{
		youtubeClient: youtubeClient,
	}

	return loader
}
