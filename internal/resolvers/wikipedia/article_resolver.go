package wikipedia

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type ArticleResolver struct {
	articleCache cache.Cache
}

func (r *ArticleResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	isWikipedia := utils.IsSubdomainOf(url, "wikipedia.org")
	isWikiArticle := strings.HasPrefix(url.Path, "/wiki/")

	return ctx, isWikipedia && isWikiArticle
}

func (r *ArticleResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	return r.articleCache.Get(ctx, url.String(), req)
}

func (r *ArticleResolver) Name() string {
	return "wikipedia:article"
}

func NewArticleResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool) *ArticleResolver {
	const endpointURL = "https://%s.wikipedia.org/api/rest_v1/page/summary/%s?redirect=false"
	articleLoader := &ArticleLoader{
		endpointURL: endpointURL,
	}

	r := &ArticleResolver{
		articleCache: cache.NewPostgreSQLCache(ctx, cfg, pool, "wikipedia:article", resolver.NewResponseMarshaller(articleLoader), 1*time.Hour),
	}

	return r
}
