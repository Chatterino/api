package twitter

import (
	"log"
	"regexp"
	"text/template"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
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

	tweetTooltipTemplate = template.Must(template.New("tweetTooltip").Parse(tweetTooltip))

	twitterUserTooltipTemplate = template.Must(template.New("twitterUserTooltip").Parse(twitterUserTooltip))

	bearerKey string

	tweetCache       = cache.New("tweets", loadTweet, 24*time.Hour)
	twitterUserCache = cache.New("twitterUsers", loadTwitterUser, 24*time.Hour)
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	if cfg.TwitterBearerToken == "" {
		log.Println("No CHATTERINO_API_TWITTER_BEARER_TOKEN specified, won't do special responses for twitter")
		return
	}
	bearerKey = cfg.TwitterBearerToken

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
