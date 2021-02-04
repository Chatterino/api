package frankerfacez

import (
	"errors"
	"regexp"
	"text/template"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	emoteAPIURL = "https://api.frankerfacez.com/v1/emote/%s"

	thumbnailFormat = "https://cdn.frankerfacez.com/emoticon/%s/4"

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

	emoteCache = cache.New("ffz_emotes", load, 1*time.Hour)

	tmpl = template.Must(template.New("frankerfacezEmoteTooltip").Parse(tooltipTemplate))

	errInvalidFrankerFaceZEmotePath = errors.New("invalid FrankerFaceZ emote path")
)

func New() (resolvers []resolver.CustomURLManager) {
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
