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

	tweetTooltipTemplate = template.Must(template.New("tweetTooltip").Parse(templateStringTweet))

	twitterUserTooltipTemplate = template.Must(template.New("twitterUserTooltip").Parse(templateStringTwitterUser))

	bearerKey string

	tweetCache       = cache.New("tweets", loadTweet, 24*time.Hour)
	twitterUserCache = cache.New("twitterUsers", loadTwitterUser, 24*time.Hour)
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	if cfg.TwitterBearerToken == "" {
		log.Println("[Config] twitter-bearer-token is missing, won't do special responses for twitter")
		return
	}
	bearerKey = cfg.TwitterBearerToken

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
