package youtube

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/staticresponse"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

type youtubeVideoTooltipData struct {
	Title         string
	ChannelTitle  string
	Duration      string
	PublishDate   string
	Views         string
	LikeCount     string
	CommentCount  string
	AgeRestricted bool
}

type youtubeStreamTooltipData struct {
	Title        string
	ChannelTitle string
	Uptime       string
	Viewers      string
	LikeCount    string
}

type VideoLoader struct {
	youtubeClient *youtubeAPI.Service
}

func (r *VideoLoader) Load(ctx context.Context, videoID string, req *http.Request) ([]byte, *int, *string, time.Duration, error) {
	log := logger.FromContext(ctx)
	youtubeVideoParts := []string{
		"statistics",
		"snippet",
		"contentDetails",
		"liveStreamingDetails",
	}

	log.Debugw("[YouTube] Get video",
		"videoID", videoID,
	)

	youtubeResponse, err := r.youtubeClient.Videos.List(youtubeVideoParts).Id(videoID).Do()
	if err != nil {
		return resolver.InternalServerErrorf("YouTube API error: %s", err)
	}

	if len(youtubeResponse.Items) == 0 {
		return staticresponse.NotFoundf("No YouTube video with the ID %s found", videoID).
			WithCacheDuration(24 * time.Hour).
			Return()
	}

	if len(youtubeResponse.Items) > 1 {
		return staticresponse.InternalServerErrorf("YouTube API returned more than %d videos", len(youtubeResponse.Items)).
			Return()
	}

	video := youtubeResponse.Items[0]

	if video.ContentDetails == nil {
		return resolver.InternalServerErrorf("YouTube video unavailable")
	}

	var tooltip bytes.Buffer

	if video.Snippet.LiveBroadcastContent == "live" {
		startTime, err := time.Parse(time.RFC3339, video.LiveStreamingDetails.ActualStartTime)
		if err != nil {
			return resolver.InternalServerErrorf("YouTube time parse error: %s", err)
		}

		data := youtubeStreamTooltipData{
			Title:        video.Snippet.Title,
			ChannelTitle: video.Snippet.ChannelTitle,
			Uptime:       humanize.Duration(time.Since(startTime)),
			Viewers:      humanize.Number(video.LiveStreamingDetails.ConcurrentViewers),
			LikeCount:    humanize.Number(video.Statistics.LikeCount),
		}

		if err := youtubeStreamTooltipTemplate.Execute(&tooltip, data); err != nil {
			return resolver.InternalServerErrorf("YouTube template error: %s", err)
		}
	} else {
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

		if err := youtubeVideoTooltipTemplate.Execute(&tooltip, data); err != nil {
			return resolver.InternalServerErrorf("YouTube template error: %s", err)
		}
	}

	thumbnail := video.Snippet.Thumbnails.Default.Url
	if video.Snippet.Thumbnails.Medium != nil {
		thumbnail = video.Snippet.Thumbnails.Medium.Url
	}

	statusCode := http.StatusOK
	contentType := "application/json"

	return buildChannelResponse(tooltip.String(), thumbnail), &statusCode, &contentType, cache.NoSpecialDur, nil
}

func NewVideoLoader(youtubeClient *youtubeAPI.Service) *VideoLoader {
	loader := &VideoLoader{
		youtubeClient: youtubeClient,
	}

	return loader
}
