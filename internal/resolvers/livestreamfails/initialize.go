package livestreamfails

import (
	"context"
	"html/template"
	"regexp"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	thumbnailFormat = "https://livestreamfails-image-prod.b-cdn.net/image/%s"

	livestreamfailsTooltipString = `<div style="text-align: left;">
{{ if .NSFW }}<li><b><span style="color: red">NSFW</span></b></li>{{ end }}
<b>{{.Title}}</b><hr>
<b>Streamer:</b> {{.StreamerName}}<br>
<b>Category:</b> {{.Category}}<br>
<b>Platform:</b> {{.Platform}}<br>
<b>Reddit score:</b> {{.RedditScore}}<br>
<b>Created:</b> {{.CreationDate}}
</div>`
)

var (
	livestreamfailsClipsTemplate = template.Must(template.New("livestreamfailsclipsTooltip").Parse(livestreamfailsTooltipString))

	pathRegex = regexp.MustCompile(`^/(?:clip|post)/([0-9]+)`)
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	const clipAPIURLFormat = "https://api.livestreamfails.com/clip/%s"
	*resolvers = append(*resolvers, NewClipResolver(ctx, cfg, pool, clipAPIURLFormat))
}
