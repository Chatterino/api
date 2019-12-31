package main

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

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

	load := func(videoID string) (interface{}, error, time.Duration) {
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

		return &LinkResolverResponse{
			Status: 200,
			Tooltip: "<div style=\"text-align: left;\"><b>" + html.EscapeString(video.Snippet.Title) +
				"</b><br><b>Channel:</b> " + html.EscapeString(video.Snippet.ChannelTitle) +
				"<br><b>Duration:</b> " + html.EscapeString(formatDuration(video.ContentDetails.Duration)) +
				"<br><b>Views:</b> " + insertCommas(strconv.FormatUint(video.Statistics.ViewCount, 10), 3) +
				"<br><span style=\"color: #2ecc71;\">" + insertCommas(strconv.FormatUint(video.Statistics.LikeCount, 10), 3) +
				" likes</span> • <span style=\"color: #e74c3c;\">" + insertCommas(strconv.FormatUint(video.Statistics.DislikeCount, 10), 3) +
				" dislikes</span></div>",
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

			apiResponse := cache.Get(videoID)
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

			apiResponse := cache.Get(videoID)
			return json.Marshal(apiResponse)
		},
	})
}
