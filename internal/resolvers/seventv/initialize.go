package seventv

import (
	"context"
	"errors"
	"html/template"
	"regexp"

	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	thumbnailFormat = "https://cdn.7tv.app/emote/%s/4x"

	tooltipTemplate = `<div style="text-align: left;">
<b>{{.Code}}</b><br>
<b>{{.Type}} SevenTV Emote</b><br>
<b>By:</b> {{.Uploader}}` +
		`{{ if .Unlisted }}` + `
<li><b><span style="color: red;">UNLISTED</span></b></li>{{ end }}
</div>`
)

var (
	seventvAPIURL = "https://api.7tv.app/v2/gql"

	errInvalidSevenTVEmotePath = errors.New("invalid SevenTV emote path")

	domains = map[string]struct{}{
		"7tv.app": {},
	}

	emotePathRegex = regexp.MustCompile(`/emotes/([a-f0-9]+)`)

	seventvEmoteTemplate = template.Must(template.New("seventvEmoteTooltip").Parse(tooltipTemplate))
)

func Initialize(ctx context.Context, cfg config.APIConfig, resolvers *[]resolver.Resolver) {
	*resolvers = append(*resolvers, NewEmoteResolver(ctx, cfg))
}
