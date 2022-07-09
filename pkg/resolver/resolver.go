package resolver

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/pkg/cache"
)

type Resolver interface {
	Check(ctx context.Context, url *url.URL) (context.Context, bool)
	Run(ctx context.Context, url *url.URL, r *http.Request) (*cache.Response, error)
	Name() string
}
