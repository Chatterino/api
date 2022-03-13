package twitch

import (
	"net/url"
	"testing"

	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
)

func TestParseClipSlug(t *testing.T) {
	c := qt.New(t)

	type parseTest struct {
		input        *url.URL
		expectedSlug string
		expectedErr  error
	}

	tests := []parseTest{}

	for _, b := range validClipBase {
		tests = append(tests, parseTest{
			input:        utils.MustParseURL(b + goodSlugV1),
			expectedSlug: goodSlugV1,
			expectedErr:  nil,
		})
		tests = append(tests, parseTest{
			input:        utils.MustParseURL(b + goodSlugV2),
			expectedSlug: goodSlugV2,
			expectedErr:  nil,
		})
	}

	for _, c := range invalidClipSlugs {
		tests = append(tests, parseTest{
			input:        utils.MustParseURL(c),
			expectedSlug: "",
			expectedErr:  errInvalidTwitchClip,
		})
	}

	for _, test := range tests {
		c.Run("", func(c *qt.C) {
			clipSlug, err := parseClipSlug(test.input)
			c.Assert(err, qt.Equals, test.expectedErr, qt.Commentf("%v", test.input))
			c.Assert(clipSlug, qt.Equals, test.expectedSlug, qt.Commentf("%v must be seen as a clip", test.input))
		})
	}
}
