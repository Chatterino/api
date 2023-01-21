package twitter

import (
	"context"
	"regexp"
	"text/template"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
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
	cfg               config.APIConfig
	tweetRegexp       = regexp.MustCompile(`^/.*\/status(?:es)?\/([^\/\?]+)`)
	twitterUserRegexp = regexp.MustCompile(`^/([^\/\?\s]+)(?:\/?$|\?.*)$`)

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

func Initialize(
	ctx context.Context,
	conf config.APIConfig,
	pool db.Pool,
	resolvers *[]resolver.Resolver,
	collageCache cache.DependentCache,
) {
	log := logger.FromContext(ctx)
	if conf.TwitterBearerToken == "" {
		log.Warnw("Twitter credentials missing, won't do special responses for Twitter")
		return
	}
	cfg = conf

	const userEndpointURLFormat = "https://api.twitter.com/2/users/by?usernames=%s&user.fields=description,profile_image_url,public_metrics"
	const tweetEndpointURLFormat = "https://api.twitter.com/2/tweets/%s?expansions=author_id,attachments.media_keys&user.fields=profile_image_url&media.fields=url&tweet.fields=created_at,public_metrics"

	*resolvers = append(
		*resolvers,
		NewTwitterResolver(
			ctx, cfg, pool, userEndpointURLFormat, tweetEndpointURLFormat, collageCache,
		),
	)
}
