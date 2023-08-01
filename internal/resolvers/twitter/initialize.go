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

	embedTweetTooltip = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<span style="white-space: pre-wrap; word-wrap: break-word;">
{{.Text}}
</span>
<span style="color: #808892;">{{.Likes}} likes&nbsp;•&nbsp;{{.Replies}} replies&nbsp;•&nbsp;{{.Timestamp}}</span>
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

	tweetTooltipTemplate      = template.Must(template.New("tweetTooltip").Parse(tweetTooltip))
	embedTweetTooltipTemplate = template.Must(template.New("tweetTooltip2").Parse(embedTweetTooltip))

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
		log.Warnw("Twitter credentials missing, using embed for Twitter")

		cfg = conf

		const tweetEndpointURLFormat = "https://cdn.syndication.twimg.com/tweet-result?id=%s&lang=en&features=tfw_timeline_list%%3A%%3Btfw_follower_count_sunset%%3Atrue%%3Btfw_tweet_edit_backend%%3Aon%%3Btfw_refsrc_session%%3Aon%%3Btfw_show_business_verified_badge%%3Aon%%3Btfw_duplicate_scribes_to_settings%%3Aon%%3Btfw_show_blue_verified_badge%%3Aon%%3Btfw_legacy_timeline_sunset%%3Atrue%%3Btfw_show_gov_verified_badge%%3Aon%%3Btfw_show_business_affiliate_badge%%3Aon%%3Btfw_tweet_edit_frontend%%3Aon"

		*resolvers = append(
			*resolvers,
			NewEmbedResolver(
				ctx, cfg, pool, tweetEndpointURLFormat, collageCache,
			),
		)

		return
	}
	cfg = conf

	const userEndpointURLFormat = "https://api.twitter.com/2/users/by?usernames=%s&user.fields=description,profile_image_url,public_metrics"
	const tweetEndpointURLFormat = "https://api.twitter.com/2/tweets/%s?expansions=author_id,attachments.media_keys&user.fields=profile_image_url&media.fields=url,preview_image_url&tweet.fields=created_at,public_metrics"

	*resolvers = append(
		*resolvers,
		NewTwitterResolver(
			ctx, cfg, pool, userEndpointURLFormat, tweetEndpointURLFormat, collageCache,
		),
	)
}
