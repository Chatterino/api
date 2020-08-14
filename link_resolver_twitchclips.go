package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dankeroni/gotwitch"
)

var noTwitchClipWithThisIDFound = &LinkResolverResponse{
	Status:  http.StatusNotFound,
	Message: "No Twitch Clip with this ID found",
}

const twitchClipsTooltip = `<div style="text-align: left;">
<b>{{.Title}}</b><hr>
<b>Channel:</b> {{.ChannelName}}<br>
<b>Views: </b> {{.Views}}</div>`

type twitchClipsTooltipData struct {
	Title       string
	ChannelName string
	Views       string
}

func init() {
	clientID, exists := os.LookupEnv("CHATTERINO_API_CACHE_TWITCH_CLIENT_ID")
	if !exists {
		log.Println("No CHATTERINO_API_CACHE_TWITCH_CLIENT_ID specified, won't do special responses for twitch clips")
		return
	}

	v5API := gotwitch.NewV5(clientID)

	tooltipTemplate, err := template.New("twitchclipsTooltip").Parse(twitchClipsTooltip)
	if err != nil {
		log.Println("Error initialization twitchclips tooltip template:", err)
		return
	}

	load := func(clipSlug string, r *http.Request) (interface{}, time.Duration, error) {
		log.Println("[TwitchClip] GET", clipSlug)
		clip, _, err := v5API.GetClip(clipSlug)
		if err != nil {
			return noTwitchClipWithThisIDFound, noSpecialDur, nil
		}

		data := twitchClipsTooltipData{
			Title:       clip.Title,
			ChannelName: clip.Broadcaster.DisplayName,
			Views:       insertCommas(strconv.FormatInt(int64(clip.Views), 10), 3),
		}

		var tooltip bytes.Buffer
		if err := tooltipTemplate.Execute(&tooltip, data); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "youtube template error " + clean(err.Error()),
			}, noSpecialDur, nil
		}

		return &LinkResolverResponse{
			Status:    200,
			Tooltip:   tooltip.String(),
			Thumbnail: clip.Thumbnails.Medium,
		}, noSpecialDur, nil
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

			apiResponse := cache.Get(clipSlug, nil)
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

			apiResponse := cache.Get(clipSlug, nil)
			return json.Marshal(apiResponse)
		},
	})
}
