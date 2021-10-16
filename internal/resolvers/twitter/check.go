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
