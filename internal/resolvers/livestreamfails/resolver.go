package livestreamfails

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

const (
	livestreamfailsAPIURL = "https://api.livestreamfails.com/clip/%s"

	thumbnailCDNFormat = "https://d2ek7gt5lc50t6.cloudfront.net/image/%s" // Hardcoded(?) cloudfront end-point
	thumbnailFormat    = "https://alpinecdn.com/v1/%s"

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

	clipCache = cache.New("livestreamfailclip", load, 1*time.Hour)

	pathRegex      = regexp.MustCompile(`/clip/([0-9]+)`)
	errInvalidPath = errors.New("invalid livestreamfails clips path")
)

func New() (resolvers []resolver.CustomURLManager) {
	// Find clips that look like https://livestreamfails.com/clip/IdHere
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			if !utils.IsSubdomainOf(url, "livestreamfails.com") {
				return false
			}

			if !pathRegex.MatchString(url.Path) {
				return false
			}

			return true
		},
		Run: func(url *url.URL) ([]byte, error) {
			pathParts := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
			clipId := pathParts[1]

			apiResponse := clipCache.Get(clipId, nil)
			return json.Marshal(apiResponse)
		},
	})

	return
}
