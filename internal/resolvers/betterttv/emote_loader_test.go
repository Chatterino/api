package betterttv

import (
	"context"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
)

func TestBuildURL(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		label     string
		baseURL   *url.URL
		emoteHash string
		expected  string
	}{
		{
			"Emote 1 real url",
			utils.MustParseURL("https://api.betterttv.net/3/emotes/"),
			"KKona",
			"https://api.betterttv.net/3/emotes/KKona",
		},
		{
			"Emote 2 real url",
			utils.MustParseURL("https://api.betterttv.net/3/emotes/"),
			"566ca04265dbbdab32ec054a",
			"https://api.betterttv.net/3/emotes/566ca04265dbbdab32ec054a",
		},
		{
			"Emote 1 fake url",
			utils.MustParseURL("http://127.0.0.1:5934/3/emotes/"),
			"566ca04265dbbdab32ec054a",
			"http://127.0.0.1:5934/3/emotes/566ca04265dbbdab32ec054a",
		},
	}

	for _, t := range tests {
		c.Run(t.label, func(c *qt.C) {
			loader := NewEmoteLoader(t.baseURL)
			actual := loader.buildURL(t.emoteHash)
			c.Assert(actual, qt.Equals, t.expected)
		})
	}
}

func TestLoad(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)
	ts := testServer()
	defer ts.Close()
	loader := NewEmoteLoader(utils.MustParseURL(ts.URL + "/3/emotes/"))

	type tc struct {
		emoteHash       string
		expectedTooltip string
		expectedMessage string
	}

	tests := []tc{
		{
			emoteHash:       "kkona",
			expectedTooltip: `<div style="text-align: left;"><b>KKona</b><br><b>Global BetterTTV Emote</b><br><b>By:</b> zneix</div>`,
		},
		{
			emoteHash:       "kkona_html",
			expectedTooltip: `<div style="text-align: left;"><b>&lt;b&gt;KKona&lt;/b&gt;</b><br><b>Global BetterTTV Emote</b><br><b>By:</b> &lt;b&gt;zneix&lt;/b&gt;</div>`,
		},
		{
			emoteHash:       "forsenga",
			expectedTooltip: `<div style="text-align: left;"><b>forsenGa</b><br><b>Shared BetterTTV Emote</b><br><b>By:</b> pajlada</div>`,
		},
		{
			emoteHash:       "forsenga_html",
			expectedTooltip: `<div style="text-align: left;"><b>&lt;b&gt;forsenGa&lt;/b&gt;</b><br><b>Shared BetterTTV Emote</b><br><b>By:</b> &lt;b&gt;pajlada&lt;/b&gt;</div>`,
		},
		{
			emoteHash:       "bad_json",
			expectedMessage: `betterttv api unmarshal error: invalid character &#39;x&#39; looking for beginning of value`,
		},
	}

	for _, test := range tests {
		c.Run(test.emoteHash, func(c *qt.C) {
			response, _, err := loader.Load(ctx, test.emoteHash, nil)

			c.Assert(err, qt.IsNil)
			c.Assert(response, qt.Not(qt.IsNil))

			cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
			c.Assert(unescapeErr, qt.IsNil)

			c.Assert(cleanTooltip, qt.Equals, test.expectedTooltip)
			c.Assert(response.Message, qt.Equals, test.expectedMessage)
		})
	}
}
