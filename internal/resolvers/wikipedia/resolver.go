package wikipedia

import (
	"errors"
	"regexp"
	"text/template"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

var (
	localeRegexp = regexp.MustCompile(`(?i)([a-z]+)\.wikipedia\.org`)
	titleRegexp  = regexp.MustCompile(`\/wiki\/(.+)`)

	wikipediaTooltipTemplate = template.Must(template.New("wikipediaTooltipTemplate").Parse(wikipediaTooltip))

	wikipediaCache = cache.New("wikipedia", load, 1*time.Hour)

	errLocaleMatch = errors.New("could not find locale from URL")
	errTitleMatch  = errors.New("could not find title from URL")
)

const (
	endpointURL = "https://%s.wikipedia.org/api/rest_v1/page/summary/%s?redirect=false"

	// TODO: Replace these with shared constants once implemented
	maxTitleLength       = 60
	maxDescriptionLength = 60
	maxExtractLength     = 200
)

func New() (resolvers []resolver.CustomURLManager) {
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
