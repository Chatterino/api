//go:generate mockgen -destination ../../mocks/mock_imgurClient.go -package=mocks . ImgurClient
// The above comment will make it so that when `go generate` is called, the command after go:generate is called in this files PWD.
// The mockgen command itself generates a mock for the ImgurClient interface in this file and stores it in the internal/mocks/ package

package imgur

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/koffeinsource/go-imgur"
)

var VALID_IMGUR_DOMAINS = []string{"imgur.com", "imgur.io"}

type ImgurClient interface {
	GetInfoFromURL(urlString string) (*imgur.GenericInfo, int, error)
}

type Resolver struct {
	imgurCache cache.Cache
}

func (r *Resolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	for _, domain := range VALID_IMGUR_DOMAINS {
		result := utils.IsSubdomainOf(url, domain)
		if result {
			return ctx, result
		}

		result = utils.IsDomain(url, domain)
		if result {
			return ctx, result
		}
	}

	return ctx, false
}

func (r *Resolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	return r.imgurCache.Get(ctx, url.String(), req)
}

func (r *Resolver) Name() string {
	return "imgur"
}

func NewResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, imgurClient ImgurClient) *Resolver {
	loader := &Loader{
		baseURL:   cfg.BaseURL,
		apiClient: imgurClient,
	}

	r := &Resolver{
		imgurCache: cache.NewPostgreSQLCache(
			ctx, cfg, pool, cache.NewPrefixKeyProvider("imgur"),
			resolver.NewResponseMarshaller(loader), cfg.ImgurCacheDuration),
	}

	return r
}
