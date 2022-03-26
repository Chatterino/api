package frankerfacez

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

var (
	data = map[string]*EmoteAPIResponse{}
)

func init() {
	data["kkona"] = &EmoteAPIResponse{
		Name: "KKona",
		Owner: EmoteAPIUser{
			DisplayName: "zneix",
		},
	}

	data["566ca04265dbbdab32ec054b"] = &EmoteAPIResponse{
		Name: "KKona",
		Owner: EmoteAPIUser{
			DisplayName: "zneix",
		},
	}

	data["kkona_html"] = &EmoteAPIResponse{
		Name: "<b>KKona</b>",
		Owner: EmoteAPIUser{
			DisplayName: "<b>zneix</b>",
		},
	}

	data["forsenga"] = &EmoteAPIResponse{
		Name: "forsenGa",
		Owner: EmoteAPIUser{
			DisplayName: "pajlada",
		},
	}

	data["forsenga_html"] = &EmoteAPIResponse{
		Name: "<b>forsenGa</b>",
		Owner: EmoteAPIUser{
			DisplayName: "<b>pajlada</b>",
		},
	}

	data["297734"] = &EmoteAPIResponse{
		Name: "pajaSx",
		Owner: EmoteAPIUser{
			DisplayName: "pajlada",
		},
	}

	data["367887"] = &EmoteAPIResponse{
		Name: "paaaajaW",
		Owner: EmoteAPIUser{
			DisplayName: "Goran42069",
		},
	}
}

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/v1/emote/{emote}", func(w http.ResponseWriter, r *http.Request) {
		emote := chi.URLParam(r, "emote")
		fmt.Println("Emote:", emote)

		type outerResponse struct {
			Emote EmoteAPIResponse `json:"emote"`
		}

		var response outerResponse

		w.Header().Set("Content-Type", "application/json")

		if emote == "696969" {
			w.Write([]byte("xD"))
		} else if e, ok := data[emote]; ok {
			response.Emote = *e
		} else {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Write(b)
	})
	return httptest.NewServer(r)
}
