package betterttv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Chatterino/api/pkg/resolver"
	qt "github.com/frankban/quicktest"
	"github.com/go-chi/chi/v5"
)

var (
	data = map[string]*EmoteAPIResponse{}
)

func init() {
	data["kkona"] = &EmoteAPIResponse{
		Code: "KKona",
		User: EmoteAPIUser{
			DisplayName: "zneix",
		},
		Global: true,
	}

	data["kkona_html"] = &EmoteAPIResponse{
		Code: "<b>KKona</b>",
		User: EmoteAPIUser{
			DisplayName: "<b>zneix</b>",
		},
		Global: true,
	}

	data["forsenga"] = &EmoteAPIResponse{
		Code: "forsenGa",
		User: EmoteAPIUser{
			DisplayName: "pajlada",
		},
	}

	data["forsenga_html"] = &EmoteAPIResponse{
		Code: "<b>forsenGa</b>",
		User: EmoteAPIUser{
			DisplayName: "<b>pajlada</b>",
		},
	}
}

func testLoadAndUnescape(c *qt.C, emote string) (cleanTooltip string) {
	iret, _, err := load(emote, nil)

	c.Assert(err, qt.IsNil)
	c.Assert(iret, qt.Not(qt.IsNil))

	response := iret.(*resolver.Response)

	c.Assert(response, qt.Not(qt.IsNil))

	cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	return cleanTooltip
}

func TestLoad(t *testing.T) {
	c := qt.New(t)
	r := chi.NewRouter()
	r.Get("/3/emotes/{emote}", func(w http.ResponseWriter, r *http.Request) {
		emote := chi.URLParam(r, "emote")

		var response *EmoteAPIResponse
		var ok bool

		w.Header().Set("Content-Type", "application/json")

		if response, ok = data[emote]; !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Write(b)
	})
	ts := httptest.NewServer(r)
	defer ts.Close()
	emoteAPIURL = ts.URL + "/3/emotes/%s"

	c.Run("Global emote", func(c *qt.C) {
		const emote = "kkona"

		const expectedTooltip = `<div style="text-align: left;"><b>KKona</b><br><b>Global BetterTTV Emote</b><br><b>By:</b> zneix</div>`

		cleanTooltip := testLoadAndUnescape(c, emote)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Global emote (HTML)", func(c *qt.C) {
		const emote = "kkona_html"

		const expectedTooltip = `<div style="text-align: left;"><b>&lt;b&gt;KKona&lt;/b&gt;</b><br><b>Global BetterTTV Emote</b><br><b>By:</b> &lt;b&gt;zneix&lt;/b&gt;</div>`

		cleanTooltip := testLoadAndUnescape(c, emote)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Shared emote", func(c *qt.C) {
		const emote = "forsenga"

		const expectedTooltip = `<div style="text-align: left;"><b>forsenGa</b><br><b>Shared BetterTTV Emote</b><br><b>By:</b> pajlada</div>`

		cleanTooltip := testLoadAndUnescape(c, emote)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Shared emote (HTML)", func(c *qt.C) {
		const emote = "forsenga_html"

		const expectedTooltip = `<div style="text-align: left;"><b>&lt;b&gt;forsenGa&lt;/b&gt;</b><br><b>Shared BetterTTV Emote</b><br><b>By:</b> &lt;b&gt;pajlada&lt;/b&gt;</div>`

		cleanTooltip := testLoadAndUnescape(c, emote)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	// c.Run("404 emote", func(c *qt.C) {
	// 	const emote = "404"

	// 	const expectedTooltip = `<div style="text-align: left;"><b>forsenGa</b><br><b>Shared BetterTTV Emote</b><br><b>By:</b> pajlada</div>`

	// 	cleanTooltip := testLoadAndUnescape(c, emote)

	// 	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	// })
}
