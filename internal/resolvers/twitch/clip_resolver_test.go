package twitch

import (
	"context"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	qt "github.com/frankban/quicktest"
)

const goodSlugV1 = "GoodSlugV1"
const goodSlugV2 = "GoodSlugV2-HVUvT7bYQnMn6nwp"

func goodSlugs(urlString string) []string {
	return []string{
		urlString + goodSlugV1,
		urlString + goodSlugV2,
	}
}

var validClips = []string{
	"https://clips.twitch.tv/" + goodSlugV1,
	"https://clips.twitch.tv/" + goodSlugV2,
	"https://twitch.tv/pajlada/clip/" + goodSlugV1,
	"https://twitch.tv/pajlada/clip/" + goodSlugV2,
	"https://twitch.tv/zneix/clip/" + goodSlugV1,
	"https://twitch.tv/zneix/clip/" + goodSlugV2,
	"https://m.twitch.tv/pajlada/clip/" + goodSlugV1,
	"https://m.twitch.tv/pajlada/clip/" + goodSlugV2,
	"https://m.twitch.tv/zneix/clip/" + goodSlugV1,
	"https://m.twitch.tv/zneix/clip/" + goodSlugV2,
	"https://m.twitch.tv/clip/" + goodSlugV1,
	"https://m.twitch.tv/clip/" + goodSlugV2,
	"https://m.twitch.tv/clip/clip/" + goodSlugV1,
	"https://m.twitch.tv/clip/clip/" + goodSlugV2,
}

var invalidClips = []string{
	"https://clips.twitch.tv/pajlada/clip/VastBitterVultureMau5",
	"https://clips.twitch.tv/",
	"https://twitch.tv/nam____________________________________________/clip/someSlugNam",
	"https://twitch.tv/supinic/clip/",
	"https://twitch.tv/pajlada/clips/VastBitterVultureMau5",
	"https://twitch.tv/zneix/clip/ImpossibleOilyAlpacaTF2John-jIlgtnSAQ52BThHhifyouseethisvivon",
	"https://twitch.tv/clip/slug",
	"https://gql.twitch.tv/VastBitterVultureMau5",
	"https://gql.twitch.tv/ThreeLetterAPI/clip/VastBitterVultureMau5",
	"https://m.twitch.tv/VastBitterVultureMau5",
	"https://m.twitch.tv/username/clip/clip/slug",
	"https://m.twitch.tv/username/notclip/slug",
}

func testCheck(ctx context.Context, resolver *ClipResolver, c *qt.C, urlString string) bool {
	u, err := url.Parse(urlString)
	c.Assert(u, qt.IsNotNil)
	c.Assert(err, qt.IsNil)

	return resolver.Check(ctx, u)
}

func TestCheck(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	resolver := &ClipResolver{}

	for _, u := range validClips {
		c.Assert(testCheck(ctx, resolver, c, u), qt.IsTrue, qt.Commentf("%v must be seen as a clip", u))
	}

	for _, u := range invalidClips {
		c.Assert(testCheck(ctx, resolver, c, u), qt.IsFalse, qt.Commentf("%v must not be seen as a clip", u))
	}
}
