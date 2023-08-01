package twitter

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type TwitterResolver struct {
	tweetCache cache.Cache
	userCache  cache.Cache
}

type EmbedResolver struct {
	tweetCache cache.Cache
}

func (r *TwitterResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if !utils.IsSubdomainOf(url, "twitter.com") {
		return ctx, false
	}

	tweetMatch := tweetRegexp.FindStringSubmatch(url.Path)
	if len(tweetMatch) == 2 && len(tweetMatch[1]) > 0 {
		return ctx, true
	}

	/* Simply matching the regex isn't enough for user pages. Pages like
	   twitter.com/explore and twitter.com/settings match the expression but do not refer
	   to a valid user page. We therefore need to check the captured name against a list
	   of known non-user pages.
	*/
	m := twitterUserRegexp.FindStringSubmatch(url.Path)
	if len(m) == 0 || len(m[1]) == 0 {
		return ctx, false
	}
	userName := strings.ToLower(m[1])

	_, notAUser := nonUserPages[userName]
	isTwitterUser := !notAUser

	return ctx, isTwitterUser
}

func (r *TwitterResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	tweetMatch := tweetRegexp.FindStringSubmatch(url.Path)
	if len(tweetMatch) == 2 && len(tweetMatch[1]) > 0 {
		tweetID := tweetMatch[1]

		return r.tweetCache.Get(ctx, tweetID, req)
	}

	userMatch := twitterUserRegexp.FindStringSubmatch(url.Path)
	if len(userMatch) == 2 && len(userMatch[1]) > 0 {
		// We always use the lowercase representation in order
		// to avoid making redundant requests.
		userName := strings.ToLower(userMatch[1])

		return r.userCache.Get(ctx, userName, req)
	}

	return nil, resolver.ErrDontHandle
}

func (r *TwitterResolver) Name() string {
	return "twitter"
}

func NewTwitterResolver(
	ctx context.Context,
	cfg config.APIConfig,
	pool db.Pool,
	userEndpointURLFormat string,
	tweetEndpointURLFormat string,
	collageCache cache.DependentCache,
) *TwitterResolver {
	tweetCacheKeyProvider := cache.NewPrefixKeyProvider("twitter:tweet")
	userCacheKeyProvider := cache.NewPrefixKeyProvider("twitter:user")

	tweetLoader := NewTweetLoader(
		cfg.BaseURL,
		cfg.TwitterBearerToken,
		tweetEndpointURLFormat,
		tweetCacheKeyProvider,
		collageCache,
		cfg.MaxThumbnailSize,
	)

	userLoader := &UserLoader{
		bearerKey:         cfg.TwitterBearerToken,
		endpointURLFormat: userEndpointURLFormat,
	}

	tweetCache := cache.NewPostgreSQLCache(
		ctx, cfg, pool, tweetCacheKeyProvider, resolver.NewResponseMarshaller(tweetLoader),
		cfg.TwitterTweetCacheDuration,
	)
	tweetCache.RegisterDependent(ctx, collageCache)

	userCache := cache.NewPostgreSQLCache(
		ctx, cfg, pool, userCacheKeyProvider, resolver.NewResponseMarshaller(userLoader),
		cfg.TwitterUserCacheDuration,
	)

	r := &TwitterResolver{
		tweetCache: tweetCache,
		userCache:  userCache,
	}

	return r
}

func (r *EmbedResolver) Check(ctx context.Context, url *url.URL) (context.Context, bool) {
	if !utils.IsSubdomainOf(url, "twitter.com") {
		return ctx, false
	}

	tweetMatch := tweetRegexp.FindStringSubmatch(url.Path)
	if len(tweetMatch) == 2 && len(tweetMatch[1]) > 0 {
		return ctx, true
	}

	return ctx, false
}

func (r *EmbedResolver) Run(ctx context.Context, url *url.URL, req *http.Request) (*cache.Response, error) {
	tweetMatch := tweetRegexp.FindStringSubmatch(url.Path)
	if len(tweetMatch) == 2 && len(tweetMatch[1]) > 0 {
		tweetID := tweetMatch[1]

		return r.tweetCache.Get(ctx, tweetID, req)
	}

	return nil, resolver.ErrDontHandle
}

func (r *EmbedResolver) Name() string {
	return "twitter-embed"
}

func NewEmbedResolver(
	ctx context.Context,
	cfg config.APIConfig,
	pool db.Pool,
	tweetEndpointURLFormat string,
	collageCache cache.DependentCache,
) *EmbedResolver {
	tweetCacheKeyProvider := cache.NewPrefixKeyProvider("twitter-embed:tweet")

	tweetLoader := NewEmbedLoader(
		cfg.BaseURL,
		tweetEndpointURLFormat,
		tweetCacheKeyProvider,
		collageCache,
		cfg.MaxThumbnailSize,
	)

	tweetCache := cache.NewPostgreSQLCache(
		ctx, cfg, pool, tweetCacheKeyProvider, resolver.NewResponseMarshaller(tweetLoader),
		cfg.TwitterTweetCacheDuration,
	)
	tweetCache.RegisterDependent(ctx, collageCache)

	r := &EmbedResolver{
		tweetCache: tweetCache,
	}

	return r
}
