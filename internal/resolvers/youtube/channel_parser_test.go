package youtube

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestGetChannelIDFromPath(t *testing.T) {
	c := qt.New(t)

	type tTest struct {
		label    string
		path     string
		expected Channel
	}

	tests := []tTest{
		{
			label: "Custom channel (/c/)",
			path:  "/c/nymnion",
			expected: Channel{
				ID:   "nymnion",
				Type: CustomChannel,
			},
		},
		{
			label: "User channel (/user/)",
			path:  "/user/nymnion",
			expected: Channel{
				ID:   "nymnion",
				Type: UserChannel,
			},
		},
		{
			label: "Identifier channel (/channel/)",
			path:  "/channel/nymnion",
			expected: Channel{
				ID:   "nymnion",
				Type: IdentifierChannel,
			},
		},
		{
			label: "Another custom channel (/CHANNELID)",
			path:  "/nymnion",
			expected: Channel{
				ID:   "nymnion",
				Type: CustomChannel,
			},
		},
		{
			label: "Invalid watch (actually a video!)",
			path:  "/watch?v=asd",
			expected: Channel{
				ID:   "",
				Type: InvalidChannel,
			},
		},
	}

	for _, test := range tests {
		c.Run(test.label, func(c *qt.C) {
			actual := getChannelFromPath(test.path)

			c.Assert(actual.ID, qt.Equals, test.expected.ID)
			c.Assert(actual.Type, qt.Equals, test.expected.Type, qt.Commentf("For path %s, got %s, should have been %s", test.path, actual.Type, test.expected.Type))
		})
	}
}
