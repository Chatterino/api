package betterttv

import (
	"context"
	"errors"
	"html/template"
	"regexp"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	thumbnailFormat = "https://cdn.betterttv.net/emote/%s/3x"

	tooltipTemplate = `<div style="text-align: left;">` +
		`<b>{{.Code}}</b><br>` +
		`<b>{{.Type}} BetterTTV Emote</b><br>` +
		`<b>By:</b> {{.Uploader}}` +
		`</div>`
)

var (
	ErrInvalidBTTVEmotePath = errors.New("invalid BetterTTV emote path")

	// BetterTTV hosts we're doing our smart things on
	domains = map[string]struct{}{
		"betterttv.com":     {},
		"www.betterttv.com": {},
	}

	emotePathRegex = regexp.MustCompile(`/emotes/([a-f0-9]+)`)

	tmpl = template.Must(template.New("betterttvEmoteTooltip").Parse(tooltipTemplate))
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	// Find links matching the BetterTTV direct emote link (e.g. https://betterttv.com/emotes/566ca06065dbbdab32ec054e)
	emoteResolver := NewEmoteResolver(ctx, cfg, pool)

	*resolvers = append(*resolvers, emoteResolver)
}
