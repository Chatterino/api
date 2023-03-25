package frankerfacez

import (
	"context"
	"errors"
	"html/template"
	"regexp"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

const (
	thumbnailFormat         = "https://cdn.frankerfacez.com/emoticon/%s/4"
	animatedThumbnailFormat = "https://cdn.frankerfacez.com/emoticon/%s/animated/4"

	tooltipTemplate = `<div style="text-align: left;">
<b>{{.Code}}</b><br>
<b>FrankerFaceZ Emote</b><br>
<b>By:</b> {{.Uploader}}</div>`
)

var (
	// FrankerFaceZ hosts we're doing our smart things on
	domains = map[string]struct{}{
		"frankerfacez.com":     {},
		"www.frankerfacez.com": {},
	}

	emotePathRegex = regexp.MustCompile(`/emoticon/([0-9]+)(-(.+)?)?$`)

	tmpl = template.Must(template.New("frankerfacezEmoteTooltip").Parse(tooltipTemplate))

	errInvalidFrankerFaceZEmotePath = errors.New("invalid FrankerFaceZ emote path")
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	emoteAPIURL := utils.MustParseURL("https://api.frankerfacez.com/v1/emote/")
	*resolvers = append(*resolvers, NewEmoteResolver(ctx, cfg, pool, emoteAPIURL))
}
