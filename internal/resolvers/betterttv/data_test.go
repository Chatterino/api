package betterttv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

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

	data["566ca04265dbbdab32ec054b"] = &EmoteAPIResponse{
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

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/3/emotes/{emote}", func(w http.ResponseWriter, r *http.Request) {
		emote := chi.URLParam(r, "emote")

		var response *EmoteAPIResponse
		var ok bool

		w.Header().Set("Content-Type", "application/json")

		if emote == "bad" {
			w.Write([]byte("xD"))
		} else if response, ok = data[emote]; !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Write(b)
	})
	return httptest.NewServer(r)
}
