package defaultresolver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
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
	"github.com/Chatterino/api/pkg/utils"
	"github.com/nicklaw5/helix"
)

type LinkResolver struct {
	customResolvers []resolver.Resolver

	linkCache      cache.Cache
	thumbnailCache cache.Cache
}

func (r *LinkResolver) HandleRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := logger.FromContext(ctx)

	log.Debugw("Handle request",
		"path", req.URL.Path,
	)
	w.Header().Set("Content-Type", "application/json")
	urlString, err := utils.UnescapeURLArgument(req, "url")
	if err != nil {
		_, err = w.Write(resolver.InvalidURL)
		if err != nil {
			log.Errorw("Error writing response",
				"error", err,
			)
		}
		return
	}

	requestUrl, err := url.Parse(urlString)
	if err != nil {
		log.Errorw("Error parsing url",
			"url", urlString,
			"error", err,
		)
		if _, err = w.Write(resolver.InvalidURL); err != nil {
			log.Errorw("Error writing response",
				"error", err,
			)
		}
	}

	for _, m := range r.customResolvers {
		if m.Check(ctx, requestUrl) {
			log.Debugw("Run url on custom resolver",
				"name", m.Name(),
				"url", requestUrl,
			)
			data, err := m.Run(ctx, requestUrl, req)

			if errors.Is(err, resolver.ErrDontHandle) {
				break
			}

			resolverHits.WithLabelValues(m.Name()).Inc()

			if err != nil {
				log.Errorw("Error in custom resolver, falling back to default",
					"name", m.Name(),
					"url", requestUrl,
					"error", err,
				)
				break
			}

			_, err = w.Write(data)
			if err != nil {
				log.Errorw("Error writing response",
					"name", m.Name(),
					"error", err,
				)
			}
			return
		}
	}

	resolverHits.WithLabelValues("default").Inc()

	response, err := r.linkCache.Get(ctx, urlString, req)
	if err != nil {
		log.Errorw("Error in default resolver",
			"url", requestUrl,
			"error", err,
		)
	} else {
		_, err = w.Write(response)
		if err != nil {
			log.Errorw("Error writing response",
				"error", err,
			)
		}
	}
}

func (r *LinkResolver) HandleThumbnailRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := logger.FromContext(ctx)

	url, err := utils.UnescapeURLArgument(req, "url")
	if err != nil {
		_, err = w.Write(resolver.InvalidURL)
		if err != nil {
			log.Errorw("Error writing response",
				"error", err,
			)
		}
		return
	}

	response, err := r.thumbnailCache.Get(ctx, url, req)

	if err != nil {
		log.Errorw("Error in thumbnail request",
			"url", url,
			"error", err,
		)
		return
	}

	_, err = w.Write(response)
	if err != nil {
		log.Errorw("Error writing response",
			"error", err,
		)
	}
}

func New(ctx context.Context, cfg config.APIConfig, helixClient *helix.Client) *LinkResolver {
	customResolvers := []resolver.Resolver{}

	// Register Link Resolvers from internal/resolvers/
	betterttv.Initialize(ctx, cfg, &customResolvers)
	discord.Initialize(ctx, cfg, &customResolvers)
	frankerfacez.Initialize(ctx, cfg, &customResolvers)
	imgur.Initialize(ctx, cfg, &customResolvers)
	livestreamfails.Initialize(ctx, cfg, &customResolvers)
	oembed.Initialize(ctx, cfg, &customResolvers)
	supinic.Initialize(ctx, cfg, &customResolvers)
	twitch.Initialize(ctx, cfg, helixClient, &customResolvers)
	twitter.Initialize(ctx, cfg, &customResolvers)
	wikipedia.Initialize(ctx, cfg, &customResolvers)
	youtube.Initialize(ctx, cfg, &customResolvers)
	seventv.Initialize(ctx, cfg, &customResolvers)

	linkLoader := &LinkLoader{
		baseURL:          cfg.BaseURL,
		maxContentLength: cfg.MaxContentLength,
		customResolvers:  customResolvers,
	}
	thumbnailLoader := &ThumbnailLoader{
		baseURL:          cfg.BaseURL,
		maxContentLength: cfg.MaxContentLength,
		enableLilliput:   cfg.EnableLilliput,
	}

	r := &LinkResolver{
		customResolvers: customResolvers,

		linkCache:      cache.NewPostgreSQLCache(ctx, cfg, "default:link", linkLoader, 10*time.Minute),
		thumbnailCache: cache.NewPostgreSQLCache(ctx, cfg, "default:thumbnail", thumbnailLoader, 10*time.Minute),
	}

	return r
}
