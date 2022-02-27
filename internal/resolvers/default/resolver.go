package defaultresolver

import (
	"errors"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/Chatterino/api/internal/logger"
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

type R struct {
	cfg config.APIConfig

	customResolvers []resolver.CustomURLManager

	defaultResolverCache          cache.Cache
	defaultResolverThumbnailCache cache.Cache

	logger logger.Logger
}

func (dr *R) HandleRequest(w http.ResponseWriter, r *http.Request) {
	dr.logger.Debugw("Handle request",
		"path", r.URL.Path,
	)
	w.Header().Set("Content-Type", "application/json")
	urlString, err := utils.UnescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(resolver.InvalidURL)
		if err != nil {
			dr.logger.Errorw("Error writing response",
				"error", err,
			)
		}
		return
	}

	requestUrl, err := url.Parse(urlString)
	if err != nil {
		dr.logger.Errorw("Error parsing url",
			"url", urlString,
			"error", err,
		)
		if _, err = w.Write(resolver.InvalidURL); err != nil {
			dr.logger.Errorw("Error writing response",
				"error", err,
			)
		}
	}

	for _, m := range dr.customResolvers {
		if m.Check(requestUrl) {
			// TODO: include custom resolver info
			dr.logger.Debugw("Run url on custom resolver",
				"url", requestUrl,
			)
			data, err := m.Run(requestUrl, r)

			if errors.Is(err, resolver.ErrDontHandle) {
				break
			}

			// TODO: Replace custom with the name of the resolver
			resolverHits.WithLabelValues("custom").Inc()

			if err != nil {
				dr.logger.Errorw("Error in custom resolver, falling back to default",
					"url", requestUrl,
					"error", err,
				)
				break
			}

			_, err = w.Write(data)
			if err != nil {
				dr.logger.Errorw("Error writing response",
					"error", err,
				)
			}
			return
		}
	}

	resolverHits.WithLabelValues("default").Inc()

	response, err := dr.defaultResolverCache.Get(urlString, r)
	if err != nil {
		dr.logger.Errorw("Error in default resolver",
			"url", requestUrl,
			"error", err,
		)
	} else {
		_, err = w.Write(response)
		if err != nil {
			dr.logger.Errorw("Error writing response",
				"error", err,
			)
		}
	}

}

func (dr *R) HandleThumbnailRequest(w http.ResponseWriter, r *http.Request) {
	url, err := utils.UnescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(resolver.InvalidURL)
		if err != nil {
			dr.logger.Errorw("Error writing response",
				"error", err,
			)
		}
		return
	}

	response, err := dr.defaultResolverThumbnailCache.Get(url, r)

	if err != nil {
		dr.logger.Errorw("Error in thumbnail request",
			"url", url,
			"error", err,
		)
		return
	}

	_, err = w.Write(response)
	if err != nil {
		dr.logger.Errorw("Error writing response",
			"error", err,
		)
	}
}

func New(cfg config.APIConfig, helixClient *helix.Client) *R {
	r := &R{
		cfg:    cfg,
		logger: cfg.Logger,
	}

	r.defaultResolverCache = cache.NewPostgreSQLCache(cfg, "linkResolver", r.load, 10*time.Minute)
	r.defaultResolverThumbnailCache = cache.NewPostgreSQLCache(cfg, "thumbnail", thumbnail.DoThumbnailRequest, 10*time.Minute)

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

	cached := stampede.Handler(512, 10*time.Second)
	imageCached := stampede.Handler(256, 2*time.Second)

	router.With(cached).Get("/link_resolver/{url}", defaultLinkResolver.HandleRequest)
	router.With(imageCached).Get("/thumbnail/{url}", defaultLinkResolver.HandleThumbnailRequest)
}
