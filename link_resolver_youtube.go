package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

const youtubeTooltip = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Channel:</b> {{.ChannelTitle}}
<br><b>Duration:</b> {{.Duration}}
<br><b>Views:</b> {{.Views}}
<br><span style="color: #2ecc71;">{{.LikeCount}} likes</span>&nbsp;•&nbsp;<span style="color: #e74c3c;">{{.DislikeCount}} dislikes</span>
</div>
`

type youtubeTooltipData struct {
	Title        string
	ChannelTitle string
	Duration     string
	Views        string
	LikeCount    string
	DislikeCount string
}

func getYoutubeVideoIDFromURL(url *url.URL) string {
	if strings.Contains(url.Path, "embed") {
		return path.Base(url.Path)
	}

	return url.Query().Get("v")
}

func getYoutubeVideoIDFromURL2(url *url.URL) string {
	return path.Base(url.Path)
}

func init() {
	apiKey, exists := os.LookupEnv("YOUTUBE_API_KEY")
	if !exists {
		log.Println("No YOUTUBE_API_KEY specified, won't do special responses for youtube")
		return
	}

	ctx := context.Background()
	youtubeClient, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Println("Error initialization youtube api client:", err)
		return
	}

	tooltipTemplate, err := template.New("youtubeTooltip").Parse(youtubeTooltip)
	if err != nil {
		log.Println("Error initialization youtube tooltip template:", err)
		return
	}

	load := func(videoID string, r *http.Request) (interface{}, error, time.Duration) {
		log.Println("[YouTube] GET", videoID)
		youtubeResponse, err := youtubeClient.Videos.List("statistics,snippet,contentDetails").Id(videoID).Do()
		if err != nil {
			return &LinkResolverResponse{
				Status:  500,
				Message: "youtube api error " + clean(err.Error()),
			}, nil, 1 * time.Hour
		}

		if len(youtubeResponse.Items) != 1 {
			return nil, errors.New("Videos response is not size 1"), noSpecialDur
		}

		video := youtubeResponse.Items[0]

		if video.ContentDetails == nil {
			return &LinkResolverResponse{Status: 500, Message: "video unavailable"}, nil, noSpecialDur
		}

		data := youtubeTooltipData{
			Title:        video.Snippet.Title,
			ChannelTitle: video.Snippet.ChannelTitle,
			Duration:     formatDuration(video.ContentDetails.Duration),
			Views:        insertCommas(strconv.FormatUint(video.Statistics.ViewCount, 10), 3),
			LikeCount:    insertCommas(strconv.FormatUint(video.Statistics.LikeCount, 10), 3),
			DislikeCount: insertCommas(strconv.FormatUint(video.Statistics.DislikeCount, 10), 3),
		}

		var tooltip bytes.Buffer
		if err := tooltipTemplate.Execute(&tooltip, data); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "youtube template error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:    http.StatusOK,
			Tooltip:   tooltip.String(),
			Thumbnail: video.Snippet.Thumbnails.Standard.Url,
		}, nil, noSpecialDur
	}

	cache := newLoadingCache("youtube", load, 24*time.Hour)

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			return strings.HasSuffix(url.Host, ".youtube.com") || url.Host == "youtube.com"
		},
		run: func(url *url.URL) ([]byte, error) {
			videoID := getYoutubeVideoIDFromURL(url)

			if videoID == "" {
				return rNoLinkInfoFound, nil
			}

			apiResponse := cache.Get(videoID, nil)
			return json.Marshal(apiResponse)
		},
	})

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			return url.Host == "youtu.be"
		},
		run: func(url *url.URL) ([]byte, error) {
			videoID := getYoutubeVideoIDFromURL2(url)

			if videoID == "" {
				return rNoLinkInfoFound, nil
			}

			apiResponse := cache.Get(videoID, nil)
			return json.Marshal(apiResponse)
		},
	})
}
