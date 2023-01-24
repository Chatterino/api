package oembed

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/dyatlov/go-oembed/oembed"
)

type Resolver struct {
	oEmbedCache cache.Cache
	oEmbed      *oembed.Oembed
}

func (r *Resolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	return ctx, r.oEmbed.FindItem(url.String()) != nil
}

func (r *Resolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	return r.oEmbedCache.Get(ctx, url.String(), req)
}

func (r *Resolver) Name() string {
	return "oembed"
}

func NewResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, data []byte) (*Resolver, error) {
	var err error
	var facebookAppAccessToken string

	if cfg.OembedFacebookAppID != "" && cfg.OembedFacebookAppSecret != "" {
		if facebookAppAccessToken, err = getFacebookAppAccessToken(cfg.OembedFacebookAppID, cfg.OembedFacebookAppSecret); err != nil {
			log.Println("[oEmbed] error loading facebook app access token", err)
		} else {
			log.Println("[oEmbed] Extra rich info loading enabled for Instagram and Facebook")
		}
	}

	oEmbed := oembed.NewOembed()
	if err := oEmbed.ParseProviders(bytes.NewReader(data)); err != nil {
		return nil, err
	}

	loader := &Loader{
		oEmbed:                 oEmbed,
		facebookAppAccessToken: facebookAppAccessToken,
	}

	r := &Resolver{
		oEmbedCache: cache.NewPostgreSQLCache(
			ctx, cfg, pool, cache.NewPrefixKeyProvider("oembed"),
			resolver.NewResponseMarshaller(loader), cfg.OembedCacheDuration,
		),
		oEmbed: oEmbed,
	}

	return r, nil
}
