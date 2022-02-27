package twitter

import (
	"log"
	"regexp"
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

	bearerKey string

	tweetCache       cache.Cache
	twitterUserCache cache.Cache
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	if cfg.TwitterBearerToken == "" {
		log.Println("[Config] twitter-bearer-token is missing, won't do special responses for twitter")
		return
	}
	bearerKey = cfg.TwitterBearerToken

	tweetCache = cache.NewPostgreSQLCache(cfg, "tweets", resolver.MarshalResponse(loadTweet), 24*time.Hour)
	twitterUserCache = cache.NewPostgreSQLCache(cfg, "twitterUsers", resolver.MarshalResponse(loadTwitterUser), 24*time.Hour)

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
