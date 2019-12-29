package main

import (
	"encoding/json"
	"html"
	"log"
	"net/url"
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
var mNoTwitchClipWithThisIDFound = mustMarshal(noTwitchClipWithThisIDFound)

func init() {
	clientID, exists := os.LookupEnv("CHATTERINO_API_CACHE_TWITCH_CLIENT_ID")
	if !exists {
		log.Println("No CHATTERINO_API_CACHE_TWITCH_CLIENT_ID specified, won't do special responses for twitch clips")
		return
	}

	v5API := gotwitch.NewV5(clientID)

	load := func(clipSlug string) (interface{}, error, time.Duration) {
		log.Println("[TwitchClip] GET", clipSlug)
		clip, _, err := v5API.GetClip(clipSlug)
		if err != nil {
			return noTwitchClipWithThisIDFound, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:  200,
			Tooltip: "<div style=\"text-align: left;\"><b>" + html.EscapeString(clip.Title) + "</b><hr><b>Channel:</b> " + html.EscapeString(clip.Broadcaster.DisplayName) + "<br><b>Views:</b> " + insertCommas(strconv.FormatInt(int64(clip.Views), 10), 3) + "</div>",
		}, nil, noSpecialDur
	}

	cache := newLoadingCache("twitchclip", load, 1*time.Hour)

	// Find clips that look like https://clips.twitch.tv/SlugHere
	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			return strings.HasSuffix(url.Host, "clips.twitch.tv")
		},
		run: func(url *url.URL) ([]byte, error) {
			pathParts := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
			clipSlug := pathParts[0]

			apiResponse := cache.Get(clipSlug)
			return json.Marshal(apiResponse)
		},
	})

	// Find clips that look like https://twitch.tv/StreamerName/clip/SlugHere
	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			if !strings.HasSuffix(url.Host, "twitch.tv") {
				return false
			}

			pathParts := strings.Split(url.Path, "/")

			return len(pathParts) >= 4 && pathParts[2] == "clip"
		},
		run: func(url *url.URL) ([]byte, error) {
			pathParts := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
			clipSlug := pathParts[2]

			apiResponse := cache.Get(clipSlug)
			return json.Marshal(apiResponse)
		},
	})
}
