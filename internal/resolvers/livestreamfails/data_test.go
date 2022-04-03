package livestreamfails

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	data = map[string]*ClipAPIResponse{}
)

func init() {
	testTime := time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC)
	// Normal clip
	data["123"] = &ClipAPIResponse{
		Category: ClipAPICategory{
			Label: "Category Label",
		},
		CreatedAt:      testTime,
		ImageID:        "asd",
		IsNSFW:         false,
		Label:          "Clip Label",
		RedditScore:    69,
		SourcePlatform: "twitch",
		Streamer: ClipAPIStreamer{
			Label: "Streamer Label",
		},
	}

	// Normal NSFW clip
	data["905"] = &ClipAPIResponse{
		Category: ClipAPICategory{
			Label: "Category Label",
		},
		CreatedAt:      testTime,
		ImageID:        "asd",
		IsNSFW:         true,
		Label:          "Clip Label",
		RedditScore:    69,
		SourcePlatform: "twitch",
		Streamer: ClipAPIStreamer{
			Label: "Streamer Label",
		},
	}
}

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/clip/{clipID}", func(w http.ResponseWriter, r *http.Request) {
		clipID := chi.URLParam(r, "clipID")

		var response *ClipAPIResponse
		var ok bool

		w.Header().Set("Content-Type", "application/json")

		if clipID == "666" {
			w.Write([]byte("xD"))
			return
		} else if clipID == "500" {
			http.Error(w, http.StatusText(500), 500)
			return
		} else if response, ok = data[clipID]; !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Write(b)
	})
	return httptest.NewServer(r)
}
