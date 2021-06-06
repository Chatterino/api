package seventv

import (
	"errors"
	"html/template"
	"regexp"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	thumbnailFormat = "https://cdn.7tv.app/emote/%s/4x"
	gqlQueryEmotes  = `{"query": "{emote(id: \"%s\") { visibility id name owner { id display_name } }}"}`

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

	seventvEmoteURLRegex = regexp.MustCompile(`7tv.app/emotes/([a-f0-9]+)`)

	emoteCache = cache.New("seventv_emotes", load, 1*time.Hour)

	seventvEmoteTemplate = template.Must(template.New("seventvEmoteTooltip").Parse(tooltipTemplate))
)

func New() (resolvers []resolver.CustomURLManager) {
	// Find links matching the SevenTV direct emote link (e.g. https://7tv.app/emotes/60b03e84b254a5e16b439128)
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
