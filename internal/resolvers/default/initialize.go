package defaultresolver

import (
	"context"
	"text/template"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/stampede"
	"github.com/nicklaw5/helix"
)

const (
	defaultTooltipString = `<div style="text-align: left;">
{{if .Title}}
<b>{{.Title}}</b><hr>
{{end}}
{{if .Description}}
<span>{{.Description}}</span><hr>
{{end}}
<b>URL:</b> {{.URL}}</div>`
)

var (
	defaultTooltip = template.Must(template.New("default_tooltip").Parse(defaultTooltipString))
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, router *chi.Mux, helixClient *helix.Client) {
	defaultLinkResolver := New(ctx, cfg, pool, helixClient)

	cached := stampede.Handler(512, 10*time.Second)
	imageCached := stampede.Handler(256, 2*time.Second)

	router.With(cached).Get("/link_resolver/{url}", defaultLinkResolver.HandleRequest)
	router.With(imageCached).Get("/thumbnail/{url}", defaultLinkResolver.HandleThumbnailRequest)
}
