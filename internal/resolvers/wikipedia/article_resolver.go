package wikipedia

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type ArticleResolver struct {
	articleCache cache.Cache
}

func (r *ArticleResolver) Check(ctx context.Context, url *url.URL) bool {
	isWikipedia := utils.IsSubdomainOf(url, "wikipedia.org")
	isWikiArticle := strings.HasPrefix(url.Path, "/wiki/")

	return isWikipedia && isWikiArticle
}

func (r *ArticleResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	return r.articleCache.Get(ctx, url.String(), req)
}

func NewArticleResolver(ctx context.Context, cfg config.APIConfig) *ArticleResolver {
	articleLoader := &ArticleLoader{}

	r := &ArticleResolver{
		articleCache: cache.NewPostgreSQLCache(ctx, cfg, "wikipedia:article", resolver.NewResponseMarshaller(articleLoader), 1*time.Hour),
	}

	return r
}
