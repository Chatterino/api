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

// getLocaleCode returns the locale code figured out from the url hostname, or "en" if none is found
func (r *ArticleResolver) getLocaleCode(u *url.URL) string {
	localeMatch := localeRegexp.FindStringSubmatch(u.Hostname())
	if len(localeMatch) != 2 {
		return "en"
	}

	return localeMatch[1]
}

// getArticleID returns the locale code figured out from the url hostname, or "en" if none is found
func (r *ArticleResolver) getArticleID(u *url.URL) (string, error) {
	titleMatch := titleRegexp.FindStringSubmatch(u.Path)
	if len(titleMatch) != 2 {
		return "", errTitleMatch
	}

	return titleMatch[1], nil
}

func (r *ArticleResolver) Check(ctx context.Context, u *url.URL) (context.Context, bool) {
	if !utils.IsSubdomainOf(u, "wikipedia.org") {
		return ctx, false
	}

	if !strings.HasPrefix(u.Path, "/wiki/") {
		return ctx, false
	}

	// Load locale code & article ID
	localeCode := r.getLocaleCode(u)
	articleID, err := r.getArticleID(u)
	if err != nil {
		return ctx, false
	}

	ctx = contextWithArticleValues(ctx, localeCode, articleID)

	// Attach locale code & article ID to context

	return ctx, true
}

func (r *ArticleResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	return r.articleCache.Get(ctx, url.String(), req)
}

func (r *ArticleResolver) Name() string {
	return "wikipedia:article"
}

func NewArticleResolver(ctx context.Context, cfg config.APIConfig, pool db.Pool, apiURL string) *ArticleResolver {
	articleLoader := &ArticleLoader{
		apiURL: apiURL,
	}

	r := &ArticleResolver{
		articleCache: cache.NewPostgreSQLCache(
			ctx, cfg, pool, cache.NewPrefixKeyProvider("wikipedia:article"),
			resolver.NewResponseMarshaller(articleLoader), 1*time.Hour,
		),
	}

	return r
}
