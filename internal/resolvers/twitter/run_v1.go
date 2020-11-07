package twitter

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/Chatterino/api/pkg/resolver"
)

func run(url *url.URL) ([]byte, error) {
	if tweetRegexp.MatchString(url.String()) {
		tweetID := getTweetIDFromURL(url)
		if tweetID == "" {
			return resolver.NoLinkInfoFound, nil
		}

		apiResponse := tweetCache.Get(tweetID, nil)
		return json.Marshal(apiResponse)
	}

	if twitterUserRegexp.MatchString(url.String()) {
		// We always use the lowercase representation in order
		// to avoid making redundant requests.
		userName := strings.ToLower(getUserNameFromUrl(url))
		if userName == "" {
			return resolver.NoLinkInfoFound, nil
		}

		apiResponse := twitterUserCache.Get(userName, nil)
		return json.Marshal(apiResponse)
	}

	return resolver.NoLinkInfoFound, nil
}
