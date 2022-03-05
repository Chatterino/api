package supinic

import (
	"context"
	"errors"
	"html/template"
	"regexp"

	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	trackListAPIURL = "https://supinic.com/api/track/detail/%d"

	tooltipTemplate = `<div style="text-align: left;">
<b>{{.Name}}</b><br>
<br>
<b>By:</b> {{.AuthorName}}<br>
<b>Track ID:</b> {{.ID}}<br>
<b>Duration:</b> {{.Duration}}<br>
<b>Tags:</b> {{.Tags}}</div>`
)

var (
	trackListTemplate = template.Must(template.New("trackListEntryTooltip").Parse(tooltipTemplate))

	errInvalidTrackPath = errors.New("invalid track list track path")

	// List of hosts that will be checked for track list paths
	trackListDomains = map[string]struct{}{
		"supinic.com": {},
	}

	trackPathRegex = regexp.MustCompile(`/track/detail/([0-9]+)`)
)

func Initialize(ctx context.Context, cfg config.APIConfig, resolvers *[]resolver.Resolver) {
	*resolvers = append(*resolvers, NewTrackResolver(ctx, cfg))
}
