package main

import (
	"encoding/json"
	"html"
	"log"
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

func init() {
	clientID, exists := os.LookupEnv("CHATTERINO_API_CACHE_TWITCH_CLIENT_ID")
	if !exists {
		log.Println("No CHATTERINO_API_CACHE_TWITCH_CLIENT_ID specified, won't do special responses for twitch clips")
		return
	}

	v5API := gotwitch.NewV5(clientID)

	load := func(clipSlug string) (interface{}, error) {
		log.Println("[TwitchClip] GET", clipSlug)
		clip, _, err := v5API.GetClip(clipSlug)
		if err != nil {
			return json.Marshal(noTwitchClipWithThisIDFound)
		}
		return &LinkResolverResponse{
			Status:  200,
			Tooltip: "<div style=\"text-align: left;\"><b>" + html.EscapeString(clip.Title) + "</b><hr><b>Channel:</b> " + html.EscapeString(clip.Broadcaster.DisplayName) + "<br><b>Views:</b> " + insertCommas(strconv.FormatInt(int64(clip.Views), 10), 3) + "</div>",
		}, nil
	}

	cache := newLoadingCache("twitchclip", load, 1*time.Hour)

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(resp *http.Response) bool {
			return strings.HasSuffix(resp.Request.URL.Host, "clips.twitch.tv")
		},
		run: func(resp *http.Response) ([]byte, error) {
			pathParts := strings.Split(strings.TrimPrefix(resp.Request.URL.Path, "/"), "/")
			clipSlug := pathParts[0]

			apiResponse := cache.Get(clipSlug)
			return json.Marshal(apiResponse)
		},
	})
}
