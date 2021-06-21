//go:generate mockgen -destination ../../mocks/mock_TwitchAPIClient.go -package=mocks . TwitchAPIClient

package twitch

import (
	"errors"
	"html/template"
	"log"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
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
	errInvalidTwitchClip = errors.New("invalid Twitch clip link")

	twitchClipsTooltip = template.Must(template.New("twitchclipsTooltip").Parse(twitchClipsTooltipString))

	clipCache = cache.New("twitchclip", load, 1*time.Hour)

	helixAPI TwitchAPIClient
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	if cfg.TwitchClientID == "" {
		log.Println("[Config] twitch_client_id is missing, won't do special responses for Twitch clips")
		return
	}

	if cfg.TwitchClientSecret == "" {
		log.Println("[Config] twitch_client_secret is missing, won't do special responses for Twitch clips")
		return
	}

	var err error

	helixAPI, err = helix.NewClient(&helix.Options{
		ClientID:     cfg.TwitchClientID,
		ClientSecret: cfg.TwitchClientSecret,
	})

	if err != nil {
		log.Fatalf("[Helix] Error initializing API client: %s", err.Error())
	}

	waitForFirstAppAccessToken := make(chan struct{})

	// Initialize methods responsible for refreshing oauth
	go initAppAccessToken(helixAPI.(*helix.Client), waitForFirstAppAccessToken)
	<-waitForFirstAppAccessToken

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
