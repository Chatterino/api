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

var (
	errInvalidTwitchClip = errors.New("invalid Twitch clip link")

	twitchClipsTooltip = template.Must(template.New("twitchclipsTooltip").Parse(twitchClipsTooltipString))

	domains = map[string]struct{}{
		"twitch.tv":       {},
		"www.twitch.tv":   {},
		"m.twitch.tv":     {},
		"clips.twitch.tv": {},
	}

	clipCache = cache.New("twitchclip", load, 1*time.Hour)

	helixAPI TwitchAPIClient
)

func New(cfg config.APIConfig, helixClient *helix.Client) (resolvers []resolver.CustomURLManager) {
	if helixClient == nil {
		log.Println("[Config] No Helix Client passed to New - won't do special responses for Twitch clips")
		return
	}

	helixAPI = helixClient

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
