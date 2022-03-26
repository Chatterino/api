package twitter

import (
	"context"
	"regexp"
	"text/template"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
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

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)
	if cfg.TwitterBearerToken == "" {
		log.Warnw("Twitter credentials missing, won't do special responses for Twitter")
		return
	}

	const userEndpointURLFormat = "https://api.twitter.com/1.1/users/show.json?screen_name=%s"
	const tweetEndpointURLFormat = "https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended"

	*resolvers = append(*resolvers, NewTwitterResolver(ctx, cfg, pool, userEndpointURLFormat, tweetEndpointURLFormat))
}
