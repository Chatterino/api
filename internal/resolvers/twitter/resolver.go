package twitter

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

type TwitterResolver struct {
	tweetCache cache.Cache
	userCache  cache.Cache
}

func (r *TwitterResolver) Check(ctx context.Context, url *url.URL) bool {
	if !utils.IsSubdomainOf(url, "twitter.com") {
		return false
	}

	isTweet := tweetRegexp.MatchString(url.String())
	if isTweet {
		return true
	}

	/* Simply matching the regex isn't enough for user pages. Pages like
	   twitter.com/explore and twitter.com/settings match the expression but do not refer
	   to a valid user page. We therefore need to check the captured name against a list
	   of known non-user pages.
	*/
	m := twitterUserRegexp.FindAllStringSubmatch(url.String(), -1)
	if len(m) == 0 || len(m[0]) == 0 {
		return false
	}
	userName := m[0][1]

	_, notAUser := nonUserPages[userName]
	isTwitterUser := !notAUser

	return isTwitterUser
}

func (r *TwitterResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	if tweetRegexp.MatchString(url.String()) {
		tweetID := getTweetIDFromURL(url)
		if tweetID == "" {
			return resolver.NoLinkInfoFound, nil
		}

		return r.tweetCache.Get(ctx, tweetID, req)
	}

	if twitterUserRegexp.MatchString(url.String()) {
		// We always use the lowercase representation in order
		// to avoid making redundant requests.
		userName := strings.ToLower(getUserNameFromUrl(url))
		if userName == "" {
			return resolver.NoLinkInfoFound, nil
		}

		return r.userCache.Get(ctx, userName, req)
	}

	// TODO: Return "do not handle" here?
	return resolver.NoLinkInfoFound, nil
}

func NewTwitterResolver(ctx context.Context, cfg config.APIConfig) *TwitterResolver {
	tweetLoader := &TweetLoader{
		bearerKey: cfg.TwitterBearerToken,
	}

	userLoader := &UserLoader{
		bearerKey: cfg.TwitterBearerToken,
	}

	r := &TwitterResolver{
		tweetCache: cache.NewPostgreSQLCache(ctx, cfg, "twitter:tweet", resolver.NewResponseMarshaller(tweetLoader), 24*time.Hour),
		userCache:  cache.NewPostgreSQLCache(ctx, cfg, "twitter:user", resolver.NewResponseMarshaller(userLoader), 24*time.Hour),
	}

	return r
}
