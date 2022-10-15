//go:generate mockgen -destination ../../mocks/mock_imgurClient.go -package=mocks . ImgurClient
// The above comment will make it so that when `go generate` is called, the command after go:generate is called in this files PWD.
// The mockgen command itself generates a mock for the ImgurClient interface in this file and stores it in the internal/mocks/ package

package imgur

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/koffeinsource/go-imgur"
)

type ImgurClient interface {
	GetInfoFromURL(urlString string) (*imgur.GenericInfo, int, error)
}

type Resolver struct {
	imgurCache cache.Cache
}

func (r *Resolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	return ctx, utils.IsSubdomainOf(url, "imgur.com")
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
			resolver.NewResponseMarshaller(loader), 1*time.Hour),
	}

	return r
}
