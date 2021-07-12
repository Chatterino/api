package defaultresolver

import (
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/Chatterino/api/internal/resolvers/betterttv"
	"github.com/Chatterino/api/internal/resolvers/discord"
	"github.com/Chatterino/api/internal/resolvers/frankerfacez"
	"github.com/Chatterino/api/internal/resolvers/imgur"
	"github.com/Chatterino/api/internal/resolvers/livestreamfails"
	"github.com/Chatterino/api/internal/resolvers/oembed"
	"github.com/Chatterino/api/internal/resolvers/seventv"
	"github.com/Chatterino/api/internal/resolvers/supinic"
	"github.com/Chatterino/api/internal/resolvers/twitch"
	"github.com/Chatterino/api/internal/resolvers/twitter"
	"github.com/Chatterino/api/internal/resolvers/wikipedia"
	"github.com/Chatterino/api/internal/resolvers/youtube"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/go-chi/chi/v5"
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

type R struct {
	cfg config.APIConfig

	customResolvers []resolver.CustomURLManager

	defaultResolverCache          *cache.Cache
	defaultResolverThumbnailCache *cache.Cache
}

func (dr *R) HandleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url, err := utils.UnescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(resolver.InvalidURL)
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	response := dr.defaultResolverCache.Get(url, r)

	_, err = w.Write(response.([]byte))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

func (dr *R) HandleThumbnailRequest(w http.ResponseWriter, r *http.Request) {
	url, err := utils.UnescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(resolver.InvalidURL)
		if err != nil {
			log.Println("Error writing thumbnail response:", err)
		}
		return
	}

	response := dr.defaultResolverThumbnailCache.Get(url, r)

	_, err = w.Write(response.([]byte))
	if err != nil {
		log.Println("Error writing thumbnail response:", err)
	}
}

func New(cfg config.APIConfig, helixClient *helix.Client) *R {
	r := &R{
		cfg: cfg,
	}

	r.defaultResolverCache = cache.New("linkResolver", r.load, 10*time.Minute)
	r.defaultResolverThumbnailCache = cache.New("thumbnail", thumbnail.DoThumbnailRequest, 10*time.Minute)

	// Register Link Resolvers from internal/resolvers/
	r.customResolvers = append(r.customResolvers, betterttv.New(cfg)...)
	r.customResolvers = append(r.customResolvers, discord.New(cfg)...)
	r.customResolvers = append(r.customResolvers, frankerfacez.New(cfg)...)
	r.customResolvers = append(r.customResolvers, imgur.New(cfg)...)
	r.customResolvers = append(r.customResolvers, livestreamfails.New(cfg)...)
	r.customResolvers = append(r.customResolvers, oembed.New(cfg)...)
	r.customResolvers = append(r.customResolvers, supinic.New(cfg)...)
	r.customResolvers = append(r.customResolvers, twitch.New(cfg, helixClient)...)
	r.customResolvers = append(r.customResolvers, twitter.New(cfg)...)
	r.customResolvers = append(r.customResolvers, wikipedia.New(cfg)...)
	r.customResolvers = append(r.customResolvers, youtube.New(cfg)...)
	r.customResolvers = append(r.customResolvers, seventv.New(cfg)...)

	return r
}

func Initialize(router *chi.Mux, cfg config.APIConfig, helixClient *helix.Client) {
	defaultLinkResolver := New(cfg, helixClient)

	router.Get("/link_resolver/{url}", defaultLinkResolver.HandleRequest)
	router.Get("/thumbnail/{url}", defaultLinkResolver.HandleThumbnailRequest)
}
