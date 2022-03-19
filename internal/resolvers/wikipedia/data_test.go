package wikipedia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/Chatterino/api/pkg/utils"
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

func testServer() *httptest.Server {
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
	return httptest.NewServer(r)
}
