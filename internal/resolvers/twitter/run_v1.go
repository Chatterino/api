package twitter

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Chatterino/api/pkg/resolver"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	if tweetRegexp.MatchString(url.String()) {
		tweetID := getTweetIDFromURL(url)
		if tweetID == "" {
			return resolver.NoLinkInfoFound, nil
		}

		return tweetCache.Get(tweetID, r)
	}

	if twitterUserRegexp.MatchString(url.String()) {
		// We always use the lowercase representation in order
		// to avoid making redundant requests.
		userName := strings.ToLower(getUserNameFromUrl(url))
		if userName == "" {
			return resolver.NoLinkInfoFound, nil
		}

		return twitterUserCache.Get(userName, r)
	}

	// TODO: Return "do not handle" here?
	return resolver.NoLinkInfoFound, nil
}
