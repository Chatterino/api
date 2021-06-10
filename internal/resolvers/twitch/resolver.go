//go:generate mockgen -destination ../../mocks/mock_TwitchAPIClient.go -package=mocks . TwitchAPIClient

package twitch

import (
	"encoding/json"
	"html/template"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/nicklaw5/helix"
)

type TwitchAPIClient interface {
	GetClips(params *helix.ClipsParams) (clip *helix.ClipsResponse, err error)
}

const (
	twitchClipsTooltipString = `<div style="text-align: left;">` +
		`<b>{{.Title}}</b><hr>` +
		`<b>Clipped by:</b> {{.AuthorName}}<br>` +
		`<b>Channel:</b> {{.ChannelName}}<br>` +
		`<b>Duration:</b> {{.Duration}}<br>` +
		`<b>Created:</b> {{.CreationDate}}<br>` +
		`<b>Views:</b> {{.Views}}` +
		`</div>`
)

var (
	twitchClipsTooltip = template.Must(template.New("twitchclipsTooltip").Parse(twitchClipsTooltipString))

	clipCache = cache.New("twitchclip", load, 1*time.Hour)

	helixAPI TwitchAPIClient
)

func New() (resolvers []resolver.CustomURLManager) {
	if config.Cfg.TwitchClientID == "" {
		log.Println("No CHATTERINO_API_TWITCH_CLIENT_ID specified, won't do special responses for Twitch clips")
		return
	}

	if config.Cfg.TwitchClientSecret == "" {
		log.Println("No CHATTERINO_API_TWITCH_CLIENT_SECRET specified, won't do special responses for Twitch clips")
		return
	}

	var err error

	helixAPI, err = helix.NewClient(&helix.Options{
		ClientID:     config.Cfg.TwitchClientID,
		ClientSecret: config.Cfg.TwitchClientSecret,
	})

	if err != nil {
		log.Fatalf("[Helix] Error initializing API client: %s", err.Error())
	}

	waitForFirstAppAccessToken := make(chan struct{})

	// Initialize methods responsible for refreshing oauth
	go initAppAccessToken(helixAPI.(*helix.Client), waitForFirstAppAccessToken)
	<-waitForFirstAppAccessToken

	// Find clips that look like https://clips.twitch.tv/SlugHere
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return utils.IsDomain(url, "clips.twitch.tv")
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
