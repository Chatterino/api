package main

import (
	"encoding/json"
	"errors"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"
)

func getYoutubeVideoIDFromURL(url *url.URL) string {
	if strings.Index(url.Path, "embed") == -1 {
		return url.Query().Get("v")
	}

	return path.Base(url.Path)
}

func init() {
	apiKey, exists := os.LookupEnv("YOUTUBE_API_KEY")
	if !exists {
		log.Println("No YOUTUBE_API_KEY specified, won't do special responses for youtube")
		return
	}

	youtubeHTTPClient := &http.Client{
		Transport: &transport.APIKey{Key: apiKey},
	}

	youtubeClient, err := youtube.New(youtubeHTTPClient)
	if err != nil {
		log.Println("Error initialization youtube api client:", err)
		return
	}

	load := func(videoID string) (interface{}, error) {
		log.Println("[YouTube] GET", videoID)
		youtubeResponse, err := youtubeClient.Videos.List("statistics,snippet,contentDetails").Id(videoID).Do()
		if err != nil {
			return nil, err
		}

		if len(youtubeResponse.Items) != 1 {
			return nil, errors.New("Videos response is not size 1")
		}

		if err != nil {
			return &LinkResolverResponse{
				Status:  500,
				Message: "youtube api error " + html.EscapeString(err.Error()),
			}, nil
		}

		video := youtubeResponse.Items[0]

		if video.ContentDetails == nil {
			return &LinkResolverResponse{Status: 500, Message: "video unavailable"}, nil
		}

		return &LinkResolverResponse{
			Status: 200,
			Tooltip: "<div style=\"text-align: left;\"><b>" + html.EscapeString(video.Snippet.Title) +
				"</b><hr><b>Channel:</b> " + html.EscapeString(video.Snippet.ChannelTitle) +
				"<br><b>Duration:</b> " + html.EscapeString(formatDuration(video.ContentDetails.Duration)) +
				"<br><b>Views:</b> " + insertCommas(strconv.FormatUint(video.Statistics.ViewCount, 10), 3) +
				"<br><b>Likes:</b> <span style=\"color: green;\">+" + insertCommas(strconv.FormatUint(video.Statistics.LikeCount, 10), 3) +
				"</span>/<span style=\"color: red;\">-" + insertCommas(strconv.FormatUint(video.Statistics.DislikeCount, 10), 3) +
				"</span></div>",
		}, nil
	}

	cache := newLoadingCache("youtube", load, 1*time.Hour)

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(resp *http.Response) bool {
			return strings.HasSuffix(resp.Request.URL.Host, ".youtube.com")
		},
		run: func(resp *http.Response) ([]byte, error) {
			videoID := getYoutubeVideoIDFromURL(resp.Request.URL)

			if videoID == "" {
				return rNoLinkInfoFound, nil
			}

			apiResponse := cache.Get(videoID)
			return json.Marshal(apiResponse)
		},
	})
}
