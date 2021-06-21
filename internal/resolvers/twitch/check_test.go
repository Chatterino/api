package twitch

import (
	"net/url"
	"testing"

	qt "github.com/frankban/quicktest"
)

func testCheck(c *qt.C, urlString string) bool {
	u, err := url.Parse(urlString)
	c.Assert(u, qt.IsNotNil)
	c.Assert(err, qt.IsNil)

	return check(u)
}

func TestCheck(t *testing.T) {
	c := qt.New(t)

	shouldCheck := []string{
		"https://clips.twitch.tv/VastBitterVultureMau5",
		"https://clips.twitch.tv/AdorableDignifiedYakFreakinStinkin-HVUvT7bYQnMn6nwp",
		"https://twitch.tv/pajlada/clip/VastBitterVultureMau5",
		"https://twitch.tv/zneix/clip/ImpossibleOilyAlpacaTF2John-jIlgtnSAQ52BThHh",
	}

	for _, u := range shouldCheck {
		c.Assert(testCheck(c, u), qt.IsTrue)
	}

	shouldNotCheck := []string{
		"https://clips.twitch.tv/pajlada/clip/VastBitterVultureMau5",
		"https://clips.twitch.tv/",
		"https://twitch.tv/nam____________________________________________/clip/someSlugNam",
		"https://twitch.tv/supinic/clip/",
		"https://twitch.tv/pajlada/clips/VastBitterVultureMau5",
		"https://twitch.tv/zneix/clip/ImpossibleOilyAlpacaTF2John-jIlgtnSAQ52BThHhifyouseethisvivon",
		"https://gql.twitch.tv/VastBitterVultureMau5",
		"https://gql.twitch.tv/ThreeLetterAPI/clip/VastBitterVultureMau5",
	}

	for _, u := range shouldNotCheck {
		c.Assert(testCheck(c, u), qt.IsFalse)
	}
}
