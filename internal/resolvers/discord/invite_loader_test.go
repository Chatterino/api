package discord

import (
	"net/url"
	"testing"

	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
)

func TestBuildURL(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		label      string
		baseURL    *url.URL
		inviteCode string
		expected   string
	}{
		{
			"Real URL 1",
			utils.MustParseURL("https://discord.com/api/v9/invites/"),
			"forsen",
			"https://discord.com/api/v9/invites/forsen",
		},
		{
			"Real URL 2",
			utils.MustParseURL("https://discord.com/api/v9/invites/"),
			"qbRE8WR",
			"https://discord.com/api/v9/invites/qbRE8WR",
		},
		{
			"Test URL 1",
			utils.MustParseURL("http://127.0.0.1:5934/api/v9/invites/"),
			"forsen",
			"http://127.0.0.1:5934/api/v9/invites/forsen",
		},
		{
			"Test URL 2",
			utils.MustParseURL("http://127.0.0.1:5934/api/v9/invites/"),
			"qbRE8WR",
			"http://127.0.0.1:5934/api/v9/invites/qbRE8WR",
		},
	}

	for _, t := range tests {
		c.Run(t.label, func(c *qt.C) {
			loader := NewInviteLoader(t.baseURL, "fakecode")
			actual := loader.buildURL(t.inviteCode)
			c.Assert(actual.String(), qt.Equals, t.expected)
		})
	}
}
