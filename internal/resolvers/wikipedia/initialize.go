package wikipedia

import (
	"context"
	"errors"
	"html/template"
	"regexp"

	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

var (
	localeRegexp = regexp.MustCompile(`(?i)([a-z]+)\.wikipedia\.org`)
	titleRegexp  = regexp.MustCompile(`\/wiki\/(.+)`)

	wikipediaTooltipTemplate = template.Must(template.New("wikipediaTooltipTemplate").Parse(wikipediaTooltip))

	errLocaleMatch = errors.New("could not find locale from URL")
	errTitleMatch  = errors.New("could not find title from URL")

	endpointURL = "https://%s.wikipedia.org/api/rest_v1/page/summary/%s?redirect=false"
)

func Initialize(ctx context.Context, cfg config.APIConfig, resolvers *[]resolver.Resolver) {
	*resolvers = append(*resolvers, NewArticleResolver(ctx, cfg))
}
