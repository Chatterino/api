package frankerfacez

import (
	"context"
	"errors"
	"html/template"
	"regexp"

	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	thumbnailFormat = "https://cdn.frankerfacez.com/emoticon/%s/4"

	tooltipTemplate = `<div style="text-align: left;">
<b>{{.Code}}</b><br>
<b>FrankerFaceZ Emote</b><br>
<b>By:</b> {{.Uploader}}</div>`
)

var (
	emoteAPIURL = "https://api.frankerfacez.com/v1/emote/%s"

	// FrankerFaceZ hosts we're doing our smart things on
	domains = map[string]struct{}{
		"frankerfacez.com":     {},
		"www.frankerfacez.com": {},
	}

	emotePathRegex = regexp.MustCompile(`/emoticon/([0-9]+)(-(.+)?)?$`)

	tmpl = template.Must(template.New("frankerfacezEmoteTooltip").Parse(tooltipTemplate))

	errInvalidFrankerFaceZEmotePath = errors.New("invalid FrankerFaceZ emote path")
)

func Initialize(ctx context.Context, cfg config.APIConfig, resolvers *[]resolver.Resolver) {
	*resolvers = append(*resolvers, NewEmoteResolver(ctx, cfg))
}
