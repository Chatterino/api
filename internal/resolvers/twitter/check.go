package twitter

import (
	"net/url"
	"strings"
)

func check(url *url.URL) bool {
	isTwitter := (strings.HasSuffix(url.Host, ".twitter.com") || url.Host == "twitter.com")

	if !isTwitter {
		return false
	}

	isTweet := tweetRegexp.MatchString(url.String())
	if isTweet {
		return true
	}

	isTwitterUser := twitterUserRegexp.MatchString(url.String())
	return isTwitterUser
}
