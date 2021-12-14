package twitch

import (
	"net/url"
	"testing"

	qt "github.com/frankban/quicktest"
)

func testParseClipSlug(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	return parseClipSlug(u)
}

func TestParseClipSlug(t *testing.T) {
	c := qt.New(t)

	for _, u := range validClips {
		clipSlug, err := testParseClipSlug(u)
		c.Assert(err, qt.IsNil, qt.Commentf("Valid clips must not error: %v", u))
		c.Assert([]string{goodSlugV1, goodSlugV2}, qt.Any(qt.Equals), clipSlug, qt.Commentf("%v must be seen as a clip", u))
	}
}
