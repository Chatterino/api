package betterttv

import (
	"net/url"
	"testing"

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
