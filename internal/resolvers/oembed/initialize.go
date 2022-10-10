package oembed

import (
	"context"
	"html/template"
	"os"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const (
	oEmbedTooltipString = `<div style="text-align: left;">
<b>{{.ProviderName}}{{ if .Title }} - {{.Title}}{{ end }}</b><hr>
{{ if .Description }}{{.Description}}{{ end }}
{{ if .AuthorName }}<br><b>Author:</b> {{.AuthorName}}{{ end }}
<br><b>URL:</b> {{.RequestedURL}}
</div>`
)

var (
	oEmbedTemplate = template.Must(template.New("oEmbedTemplateTooltip").Parse(oEmbedTooltipString))
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)

	data, err := os.ReadFile(cfg.OembedProvidersPath)

	if err != nil {
		log.Warnw("[oEmbed] No providers.json file found, won't do oEmbed parsing")
		return
	}

	resolver, err := NewResolver(ctx, cfg, pool, data)
	if err != nil {
		log.Warnw("[oEmbed] Error parsing providers.json file",
			"error", err,
		)
	}

	*resolvers = append(*resolvers, resolver)
}
