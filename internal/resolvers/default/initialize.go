package defaultresolver

import (
	"context"
	"log/slog"
	"text/template"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/stampede"
	memcache "github.com/goware/cachestore-mem"
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

var defaultTooltip = template.Must(template.New("default_tooltip").Parse(defaultTooltipString))

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, router *chi.Mux, helixClient *helix.Client) {
	// Ignored hosts can be added here at request of the hoster
	ignoredHosts := map[string]struct{}{}

	defaultLinkResolver := New(ctx, cfg, pool, helixClient, ignoredHosts)

	imageCache, err := memcache.NewBackend(256)
	if err != nil {
		panic(err)
	}

	imageCached := stampede.Handler(slog.Default(), imageCache, 2*time.Second)

	genValueCache, err := memcache.NewBackend(256)
	if err != nil {
		panic(err)
	}
	generatedValuesCached := stampede.Handler(slog.Default(), genValueCache, 2*time.Second)

	// TODO: Make the max age headers be applied based on the resolved link's configured cache timer
	router.With(cache.MaxAgeHeaders(time.Minute*10)).Get("/link_resolver/{url}", defaultLinkResolver.HandleRequest)
	router.With(cache.MaxAgeHeaders(time.Minute*10), imageCached).Get("/thumbnail/{url}", defaultLinkResolver.HandleThumbnailRequest)
	router.With(generatedValuesCached).Get("/generated/{url}", defaultLinkResolver.HandleGeneratedValueRequest)
}
