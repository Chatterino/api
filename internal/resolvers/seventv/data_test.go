package seventv

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

var (
	emotes = map[string]EmoteAPIResponse{}
)

func init() {
	// Private emote: Pajawalk
	emotes["604281c81ae70f000d47ffd9"] = EmoteAPIResponse{
		Data: EmoteAPIResponseData{
			Emote: &EmoteAPIEmote{
				ID:         "604281c81ae70f000d47ffd9",
				Name:       "Pajawalk",
				Visibility: EmoteVisibilityPrivate,
				Owner: EmoteAPIUser{
					ID:          "603d10e496832ffa787ca53c",
					DisplayName: "durado_",
				},
			},
		},
	}

	// Hidden emote: Bedge
	emotes["60ae8d9ff39a7552b658b60d"] = EmoteAPIResponse{
		Data: EmoteAPIResponseData{
			Emote: &EmoteAPIEmote{
				ID:         "60ae8d9ff39a7552b658b60d",
				Name:       "Bedge",
				Visibility: EmoteVisibilityPrivate | EmoteVisibilityHidden,
				Owner: EmoteAPIUser{
					ID:          "605394d9b4d31e459ff05f40",
					DisplayName: "Paruna",
				},
			},
		},
	}

	// Global emote: FeelsOkayMan
	emotes["6042998c1d4963000d9dae34"] = EmoteAPIResponse{
		Data: EmoteAPIResponseData{
			Emote: &EmoteAPIEmote{
				ID:         "6042998c1d4963000d9dae34",
				Name:       "FeelsOkayMan",
				Visibility: EmoteVisibilityGlobal,
				Owner: EmoteAPIUser{
					ID:          "603bb6a596832ffa78e7b27b",
					DisplayName: "MegaKill3",
				},
			},
		},
	}

	// No visiblity tag emote: monkaE
	emotes["603cb219c20d020014423c34"] = EmoteAPIResponse{
		Data: EmoteAPIResponseData{
			Emote: &EmoteAPIEmote{
				ID:   "603cb219c20d020014423c34",
				Name: "monkaE",
				// No visibility, should default to shared
				Owner: EmoteAPIUser{
					ID:          "6042058896832ffa785800fe",
					DisplayName: "Zhark",
				},
			},
		},
	}

	// No emote
	emotes["f0f0f0"] = EmoteAPIResponse{
		Data: EmoteAPIResponseData{
			Emote: nil,
		},
	}
}

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Post("/v2/gql", func(w http.ResponseWriter, r *http.Request) {
		type gqlQuery struct {
			Query     string            `json:"string"`
			Variables map[string]string `json:"variables"`
		}

		var q gqlQuery

		xd, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(xd, &q)
		if err != nil {
			panic(err)
		}

		emoteID := q.Variables["id"]

		if emoteID == "bad" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("xd"))
			return
		} else if response, ok := emotes[emoteID]; ok {
			b, _ := json.Marshal(&response)

			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
			return
		} else {
			http.Error(w, http.StatusText(404), 404)
			return
		}
	})
	return httptest.NewServer(r)
}
