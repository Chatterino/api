//go:generate mockgen -destination ../../mocks/mock_TwitchAPIClient.go -package=mocks . TwitchAPIClient

package twitch

import (
	"context"
	"errors"
	"html/template"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/nicklaw5/helix"
)

type TwitchAPIClient interface {
	GetClips(params *helix.ClipsParams) (clip *helix.ClipsResponse, err error)
	GetUsers(params *helix.UsersParams) (clip *helix.UsersResponse, err error)
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

	twitchUserTooltipString = `<div style="text-align: left;">` +
		`<b>{{.Name}}</b><br>` +
		`{{.Description}}<br>` +
		`<b>Created:</b> {{.CreatedAt}}` +
		`</div>`
)

var (
	errInvalidTwitchClip = errors.New("invalid Twitch clip link")

	twitchClipsTooltip = template.Must(template.New("twitchclipsTooltip").Parse(twitchClipsTooltipString))
	twitchUserTooltip  = template.Must(template.New("twitchUserTooltip").Parse(twitchUserTooltipString))

	domains = map[string]struct{}{
		"twitch.tv":       {},
		"www.twitch.tv":   {},
		"m.twitch.tv":     {},
		"clips.twitch.tv": {},
	}
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, helixClient TwitchAPIClient, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)

	if utils.IsInterfaceNil(helixClient) {
		log.Warnw("[Config] Twitch credentials missing, won't do special responses for Twitch")
		return
	}

	*resolvers = append(*resolvers, NewUserResolver(ctx, cfg, pool, helixClient))
	*resolvers = append(*resolvers, NewClipResolver(ctx, cfg, pool, helixClient))
}
