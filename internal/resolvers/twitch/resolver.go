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
	"github.com/nicklaw5/helix"
	// "github.com/dankeroni/gotwitch"
)

const (
	twitchClipsTooltipString = `<div style="text-align: left;">
<b>{{.Title}}</b><hr>
<b>Channel:</b> {{.ChannelName}}<br>
<b>Duration:</b> {{.Duration}}<br>
<b>Created:</b> {{.CreationDate}}<br>
<b>Views: </b> {{.Views}}</div>`
)

var (
	twitchClipsTooltip = template.Must(template.New("twitchclipsTooltip").Parse(twitchClipsTooltipString))

	clipCache = cache.New("twitchclip", load, 1*time.Hour)

	// v5API *gotwitch.TwitchAPI

	helixAPI *helix.Client
)

func New() (resolvers []resolver.CustomURLManager) {
	clientID, existsClientID := os.LookupEnv("CHATTERINO_API_TWITCH_CLIENT_ID")
	clientSecret, existsClientSecret := os.LookupEnv("CHATTERINO_API_TWITCH_CLIENT_SECRET")

	if !existsClientID {
		log.Println("No CHATTERINO_API_TWITCH_CLIENT_ID specified, won't do special responses for Twitch clips")
		return
	}

	if !existsClientSecret {
		log.Println("No CHATTERINO_API_TWITCH_CLIENT_SECRET specified, won't do special responses for Twitch clips")
		return
	}

	// v5API = gotwitch.NewV5(clientID)
	helixAPI, err := helix.NewClient(&helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	if err != nil {
		log.Fatalf("[HELIX] Error initializing API client: %s", err.Error())
	}

	response, err := helixAPI.RequestAppAccessToken([]string{})

	if err != nil {
		log.Fatalf("[HELIX] Error requesting app access token: %s , \n %s", err.Error(), response.Error)
	}

	log.Printf("Requested access token %s response: %d, expires in: %d", response.Data.AccessToken, response.StatusCode, response.Data.ExpiresIn)
	helixAPI.SetAppAccessToken(response.Data.AccessToken)

	// // Refresh app access token every 24 hours
	// ticker := time.NewTicker(24 * time.Hour)

	// for range ticker.C {
	// 	response, err := helixAPI.RequestAppAccessToken([]string{})
	// 	if err != nil {
	// 		log.Print(err.Error())
	// 		continue
	// 	}
	// 	log.Printf("Requested access token from routine, response: %d, expires in: %d", response.StatusCode, response.Data.ExpiresIn)

	// 	helixAPI.SetAppAccessToken(response.Data.AccessToken)
	// }

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
