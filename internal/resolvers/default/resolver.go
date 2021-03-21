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
	"github.com/Chatterino/api/internal/resolvers/supinic"
	"github.com/Chatterino/api/internal/resolvers/twitch"
	"github.com/Chatterino/api/internal/resolvers/twitter"
	"github.com/Chatterino/api/internal/resolvers/wikipedia"
	"github.com/Chatterino/api/internal/resolvers/youtube"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/go-chi/chi/v5"
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
	baseURL string

	customResolvers []resolver.CustomURLManager

	defaultResolverCache *cache.Cache
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

func New(baseURL string) *R {
	r := &R{
		baseURL: baseURL,
	}

	r.defaultResolverCache = cache.New("linkResolver", r.load, time.Duration(10)*time.Minute)

	// Register Link Resolvers from internal/resolvers/
	r.customResolvers = append(r.customResolvers, betterttv.New()...)
	r.customResolvers = append(r.customResolvers, frankerfacez.New()...)
	r.customResolvers = append(r.customResolvers, twitter.New()...)
	r.customResolvers = append(r.customResolvers, discord.New()...)
	r.customResolvers = append(r.customResolvers, youtube.New()...)
	r.customResolvers = append(r.customResolvers, supinic.New()...)
	r.customResolvers = append(r.customResolvers, twitch.New()...)
	r.customResolvers = append(r.customResolvers, imgur.New()...)
	r.customResolvers = append(r.customResolvers, wikipedia.New()...)
	r.customResolvers = append(r.customResolvers, livestreamfails.New()...)

	return r
}

func Initialize(router *chi.Mux, baseURL string) {
	defaultLinkResolver := New(baseURL)

	router.Get("/link_resolver/{url}", defaultLinkResolver.HandleRequest)
}
