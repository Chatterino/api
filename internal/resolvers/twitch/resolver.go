package twitch

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/dankeroni/gotwitch"
)

const (
	twitchClipsTooltipString = `<div style="text-align: left;">
<b>{{.Title}}</b><hr>
<b>Clipped by:</b> {{.AuthorName}}<br>
<b>Channel:</b> {{.ChannelName}}<br>
<b>Duration:</b> {{.Duration}}<br>
<b>Created:</b> {{.CreationDate}}<br>
<b>Views:</b> {{.Views}}</div>`
)

var (
	twitchClipsTooltip = template.Must(template.New("twitchclipsTooltip").Parse(twitchClipsTooltipString))

	clipCache = cache.New("twitchclip", load, 1*time.Hour)

	v5API *gotwitch.TwitchAPI
)

func New() (resolvers []resolver.CustomURLManager) {
	clientID, exists := os.LookupEnv("CHATTERINO_API_CACHE_TWITCH_CLIENT_ID")
	if !exists {
		log.Println("No CHATTERINO_API_CACHE_TWITCH_CLIENT_ID specified, won't do special responses for twitch clips")
		return
	}

	v5API = gotwitch.NewV5(clientID)

	// Find clips that look like https://clips.twitch.tv/SlugHere
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return strings.HasSuffix(url.Host, "clips.twitch.tv")
		},
		Run: func(url *url.URL) ([]byte, error) {
			pathParts := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
			clipSlug := pathParts[0]

			apiResponse := clipCache.Get(clipSlug, nil)
			return json.Marshal(apiResponse)
		},
	})

	// Find clips that look like https://twitch.tv/StreamerName/clip/SlugHere
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			if !strings.HasSuffix(url.Host, "twitch.tv") {
				return false
			}

			pathParts := strings.Split(url.Path, "/")

			return len(pathParts) >= 4 && pathParts[2] == "clip"
		},
		Run: func(url *url.URL) ([]byte, error) {
			pathParts := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
			clipSlug := pathParts[2]

			apiResponse := clipCache.Get(clipSlug, nil)
			return json.Marshal(apiResponse)
		},
	})

	return
}
