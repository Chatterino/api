package main

import (
	"encoding/json"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dankeroni/gotwitch"
)

var noTwitchClipWithThisIDFound = &LinkResolverResponse{
	Status:  404,
	Message: "No Twitch Clip with this ID found",
}

func getTwitchClip(statusCode int, v5API *gotwitch.TwitchAPI, clipSlug string) func() (interface{}, error) {
	return func() (interface{}, error) {
		clip, _, err := v5API.GetClip(clipSlug)
		if err != nil {
			return json.Marshal(noTwitchClipWithThisIDFound)
		}
		return &LinkResolverResponse{
			Status:  statusCode,
			Tooltip: "<div style=\"text-align: left;\"><b>" + html.EscapeString(clip.Title) + "</b><hr><b>Channel:</b> " + html.EscapeString(clip.Broadcaster.DisplayName) + "<br><b>Views:</b> " + insertCommas(strconv.FormatInt(int64(clip.Views), 10), 3) + "</div>",
		}, nil
	}
}

func init() {
	clientID, exists := os.LookupEnv("CHATTERINO_API_CACHE_TWITCH_CLIENT_ID")
	if !exists {
		return
	}

	v5API := gotwitch.NewV5(clientID)

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(resp *http.Response) bool {
			return strings.HasSuffix(resp.Request.URL.Host, "clips.twitch.tv")
		},
		run: func(resp *http.Response) ([]byte, error) {
			pathParts := strings.Split(strings.TrimPrefix(resp.Request.URL.Path, "/"), "/")
			clipSlug := pathParts[0]
			twitchClipResponse := cacheGetOrSet("twitchclip:"+clipSlug, 1*time.Hour, getTwitchClip(resp.StatusCode, v5API, clipSlug))
			return json.Marshal(twitchClipResponse)
		},
	})
}
