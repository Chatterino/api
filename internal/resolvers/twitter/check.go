package twitter

import (
	"net/url"

	"github.com/Chatterino/api/pkg/utils"
)

func check(url *url.URL) bool {
	if !utils.IsSubdomainOf(url, "twitter.com") {
		return false
	}

	isTweet := tweetRegexp.MatchString(url.String())
	if isTweet {
		return true
	}

	isTwitterUser := twitterUserRegexp.MatchString(url.String())
	return isTwitterUser
}
