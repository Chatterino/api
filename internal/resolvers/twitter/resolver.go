package twitter

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

const (
	tweetTooltip = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<span style="white-space: pre-wrap; word-wrap: break-word;">
{{.Text}}
</span>
<span style="color: #808892;">{{.Likes}} likes&nbsp;•&nbsp;{{.Retweets}} retweets&nbsp;•&nbsp;{{.Timestamp}}</span>
</div>
`

	twitterUserTooltip = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<span style="white-space: pre-wrap; word-wrap: break-word;">
{{.Description}}
</span>
<span style="color: #808892;">{{.Followers}} followers</span>
</div>
`
)

var (
	tweetRegexp       = regexp.MustCompile(`(?i)\/.*\/status(?:es)?\/([^\/\?]+)`)
	twitterUserRegexp = regexp.MustCompile(`(?i)twitter\.com\/([^\/\?\s]+)(\/?$|(\?).*)`)

	/* These routes refer to non-user pages. If the capture group in twitterUserRegexp
	   matches any of these names, we must not resolve it as a Twitter user link.

	   The pages are listed alphabetically. They were sourced by simply looking around the
	   Twitter web page. AFAIK, there is no resource describing these "special" routes.
	*/
	nonUserPages = utils.SetFromSlice([]interface{}{
		"compose",
		"explore",
		"home",
		"logout",
		"messages",
		"notifications",
		"search",
		"settings",
		"tos",
		"privacy",
	})

	tweetTooltipTemplate = template.Must(template.New("tweetTooltip").Parse(tweetTooltip))

	twitterUserTooltipTemplate = template.Must(template.New("twitterUserTooltip").Parse(twitterUserTooltip))
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

func NewTwitterResolver(ctx context.Context, cfg config.APIConfig) (*TwitterResolver, error) {
	if cfg.TwitterBearerToken == "" {
		return nil, errors.New("twitter-bearer-token is missing, won't do special responses for Twitter")
	}

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

	return r, nil
}

func Initialize(ctx context.Context, cfg config.APIConfig, resolvers *[]resolver.Resolver) {
	resolver, err := NewTwitterResolver(ctx, cfg)
	if err != nil {
		return
	}

	*resolvers = append(*resolvers, resolver)
}
