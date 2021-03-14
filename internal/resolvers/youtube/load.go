package youtube

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

func parseDuration(dur string) string {
	dur = strings.ToLower(dur)
	dur = strings.Replace(dur, "pt", "", 1)
	d, _ := time.ParseDuration(dur)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func load(videoID string, r *http.Request) (interface{}, time.Duration, error) {
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

	data := youtubeTooltipData{
		Title:        video.Snippet.Title,
		ChannelTitle: video.Snippet.ChannelTitle,
		Duration:     parseDuration(video.ContentDetails.Duration),
		PublishDate:  humanize.Date("02 Jan 2006", video.Snippet.PublishedAt),
		Views:        humanize.Number(video.Statistics.ViewCount),
		LikeCount:    humanize.Number(video.Statistics.LikeCount),
		DislikeCount: humanize.Number(video.Statistics.DislikeCount),
	}

	var tooltip bytes.Buffer
	if err := youtubeTooltipTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "youtube template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	thumbnail := video.Snippet.Thumbnails.Default.Url
	if video.Snippet.Thumbnails.Standard != nil {
		thumbnail = video.Snippet.Thumbnails.Standard.Url
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnail,
	}, cache.NoSpecialDur, nil
}
