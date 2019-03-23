package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"
)

func getYoutubeVideo(youtubeClient *youtube.Service, videoID string) (*youtube.Video, error) {
	youtubeResponse, err := youtubeClient.Videos.List("statistics,snippet,contentDetails").Id(videoID).Do()
	if err != nil {
		return nil, err
	}

	if len(youtubeResponse.Items) != 1 {
		return nil, errors.New("Videos response is not size 1")
	}

	return youtubeResponse.Items[0], nil
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

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(resp *http.Response) bool {
			return strings.HasSuffix(resp.Request.URL.Host, ".youtube.com")
		},
		run: func(resp *http.Response) ([]byte, error) {
			url := resp.Request.URL
			videoID := ""

			if strings.Index(url.Path, "embed") == -1 {
				videoID = url.Query().Get("v")
			} else {
				videoID = path.Base(url.Path)
			}

			if videoID == "" {
				return json.Marshal(noLinkInfoFound)
			}

			youtubeResponse := cacheGetOrSet("youtube:"+videoID, 1*time.Hour, func() (interface{}, error) {
				video, err := getYoutubeVideo(youtubeClient, videoID)
				if err != nil {
					return &LinkResolverResponse{Status: 500, Message: "youtube api error " + err.Error()}, nil
				}

				fmt.Println("Doing YouTube API Request on", videoID)
				return &LinkResolverResponse{
					Status: resp.StatusCode,
					Tooltip: "<div style=\"text-align: left;\"><b>" + html.EscapeString(video.Snippet.Title) +
						"</b><hr><b>Channel:</b> " + html.EscapeString(video.Snippet.ChannelTitle) +
						"<br><b>Duration:</b> " + html.EscapeString(formatDuration(video.ContentDetails.Duration)) +
						"<br><b>Views:</b> " + insertCommas(strconv.FormatUint(video.Statistics.ViewCount, 10), 3) +
						"<br><b>Likes:</b> <span style=\"color: green;\">+" + insertCommas(strconv.FormatUint(video.Statistics.LikeCount, 10), 3) +
						"</span>/<span style=\"color: red;\">-" + insertCommas(strconv.FormatUint(video.Statistics.DislikeCount, 10), 3) +
						"</span></div>",
				}, nil
			})

			return json.Marshal(youtubeResponse)
		},
	})
}
