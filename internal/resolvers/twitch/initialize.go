//go:generate mockgen -destination ../../mocks/mock_TwitchAPIClient.go -package=mocks . TwitchAPIClient

package twitch

import (
	"context"
	"errors"
	"html/template"
	"reflect"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
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

	domains = map[string]struct{}{
		"twitch.tv":       {},
		"www.twitch.tv":   {},
		"m.twitch.tv":     {},
		"clips.twitch.tv": {},
	}
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, helixClient TwitchAPIClient, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)

	checkNil := reflect.ValueOf(helixClient)

	if !checkNil.IsValid() || checkNil.IsNil() {
		log.Warnw("[Config] Twitch credentials missing, won't do special responses for Twitch")
		return
	}

	*resolvers = append(*resolvers, NewClipResolver(ctx, cfg, pool, helixClient))
}
