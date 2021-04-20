package wikipedia

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/go-chi/chi/v5"
)

var (
	wikiData = map[string]*wikipediaAPIResponse{}
)

func init() {
	wikiData["en_test"] = &wikipediaAPIResponse{
		Titles: wikipediaAPITitles{
			Normalized: "Test title",
		},
		Extract:     "Test extract",
		Thumbnail:   nil,
		Description: utils.StringPtr("Test description"),
	}

	wikiData["en_test_html"] = &wikipediaAPIResponse{
		Titles: wikipediaAPITitles{
			Normalized: "<b>Test title</b>",
		},
		Extract:     "<b>Test extract</b>",
		Thumbnail:   nil,
		Description: utils.StringPtr("<b>Test description</b>"),
	}

	wikiData["en_test_no_description"] = &wikipediaAPIResponse{
		Titles: wikipediaAPITitles{
			Normalized: "Test title",
		},
		Extract:     "Test extract",
		Thumbnail:   nil,
		Description: nil,
	}
}

func testLoadAndUnescape(c *qt.C, locale, page string) (cleanTooltip string) {
	urlString := fmt.Sprintf("https://%s.wikipedia.org/wiki/%s", locale, page)
	iret, _, err := load(urlString, nil)

	c.Assert(err, qt.IsNil)
	c.Assert(iret, qt.Not(qt.IsNil))

	response := iret.(response)

	c.Assert(response, qt.Not(qt.IsNil))
	c.Assert(response.err, qt.IsNil)

	c.Assert(response.resolverResponse, qt.Not(qt.IsNil))

	cleanTooltip, unescapeErr := url.PathUnescape(response.resolverResponse.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	return cleanTooltip
}

func TestLoad(t *testing.T) {
	c := qt.New(t)
	r := chi.NewRouter()
	r.Get("/api/rest_v1/page/summary/{locale}/{page}", func(w http.ResponseWriter, r *http.Request) {
		locale := chi.URLParam(r, "locale")
		page := chi.URLParam(r, "page")

		var response *wikipediaAPIResponse
		var ok bool

		if response, ok = wikiData[locale+"_"+page]; !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
	ts := httptest.NewServer(r)
	defer ts.Close()
	endpointURL = ts.URL + "/api/rest_v1/page/summary/%s/%s"

	c.Run("Normal page", func(c *qt.C) {
		const locale = "en"
		const page = "test"

		const expectedTooltip = `<div style="text-align: left;"><b>Test title&nbsp;•&nbsp;Test description</b><br>Test extract</div>`

		cleanTooltip := testLoadAndUnescape(c, locale, page)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal page (HTML)", func(c *qt.C) {
		const locale = "en"
		const page = "test_html"

		const expectedTooltip = `<div style="text-align: left;"><b>&lt;b&gt;Test title&lt;/b&gt;&nbsp;•&nbsp;&lt;b&gt;Test description&lt;/b&gt;</b><br>&lt;b&gt;Test extract&lt;/b&gt;</div>`

		cleanTooltip := testLoadAndUnescape(c, locale, page)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal page (No description)", func(c *qt.C) {
		const locale = "en"
		const page = "test_no_description"

		const expectedTooltip = `<div style="text-align: left;"><b>Test title</b><br>Test extract</div>`

		cleanTooltip := testLoadAndUnescape(c, locale, page)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	// c.Run("Nonexistant page", func(c *qt.C) {
	// 	const locale = "en"
	// 	const page = "404"

	// 	const expectedTooltip = `404`

	// 	cleanTooltip := testLoadAndUnescape(c, locale, page)

	// 	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	// })
}
