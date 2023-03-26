package seventv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

var (
	emotes = map[string]EmoteModel{}
)

func init() {
	// Private emote: Pajawalk
	emotes["604281c81ae70f000d47ffd9"] = EmoteModel{
		ID:     "604281c81ae70f000d47ffd9",
		Name:   "Pajawalk",
		Flags:  EmoteFlagsPrivate,
		Listed: true,
		Host: ImageHost{
			URL:   "//cdn.7tv.app/emote/604281c81ae70f000d47ffd9",
			Files: []ImageFile{{Name: "best.avif", Width: 100, Height: 30, Format: ImageFormatAVIF}, {Name: "best.webp", Width: 90, Height: 28, Format: ImageFormatWEBP}},
		},
		Owner: UserPartialModel{
			ID:          "603d10e496832ffa787ca53c",
			DisplayName: "durado_",
		},
	}

	// Unlisted emote: Bedge
	emotes["60ae8d9ff39a7552b658b60d"] = EmoteModel{
		ID:     "60ae8d9ff39a7552b658b60d",
		Name:   "Bedge",
		Flags:  0,
		Listed: false,
		Host: ImageHost{
			URL:   "//cdn.7tv.app/emote/60ae8d9ff39a7552b658b60d",
			Files: []ImageFile{{Name: "best.webp", Width: 90, Height: 28, Format: ImageFormatWEBP}},
		},
		Owner: UserPartialModel{
			ID:          "605394d9b4d31e459ff05f40",
			DisplayName: "Paruna",
		},
	}

	// Regular emote: monkaS
	emotes["603cb219c20d020014423c34"] = EmoteModel{
		ID:     "603cb219c20d020014423c34",
		Name:   "monkaS",
		Flags:  0,
		Listed: true,
		Host: ImageHost{
			URL:   "https://cdn.7tv.app/emote/603cb219c20d020014423c34",
			Files: []ImageFile{{Name: "1x.webp", Width: 28, Height: 28, Format: ImageFormatWEBP}, {Name: "best.webp", Width: 128, Height: 128, Format: ImageFormatWEBP}},
		},
		Owner: UserPartialModel{
			ID:          "6042058896832ffa785800fe",
			DisplayName: "Zhark",
		},
	}

	// Regular emote, no webp images: Hmm
	emotes["60ae3e54259ac5a73e56a426"] = EmoteModel{
		ID:     "60ae3e54259ac5a73e56a426",
		Name:   "Hmm",
		Flags:  0,
		Listed: true,
		Host: ImageHost{
			URL:   "https://cdn.7tv.app/emote/60ae3e54259ac5a73e56a426",
			Files: []ImageFile{{Name: "jebaited.webp", Width: 128, Height: 128, Format: ImageFormatAVIF}},
		},
		Owner: UserPartialModel{
			ID:          "60772a85a807bed00612d1ee",
			DisplayName: "lnsc",
		},
	}

	// Private, unlisted emote: Okayge
	emotes["60bcb44f7229037ee386d1ab"] = EmoteModel{
		ID:     "60bcb44f7229037ee386d1ab",
		Name:   "Okayge",
		Flags:  EmoteFlagsPrivate,
		Listed: false,
		Host: ImageHost{
			URL:   "//cdn.7tv.app/emote/60bcb44f7229037ee386d1ab",
			Files: []ImageFile{{Name: "1x.webp", Width: 28, Height: 28, Format: ImageFormatWEBP}, {Name: "best.webp", Width: 128, Height: 128, Format: ImageFormatWEBP}, {Name: "2x.webp", Width: 42, Height: 42, Format: ImageFormatWEBP}},
		},
		Owner: UserPartialModel{
			ID:          "60aeabfff6a2c3b332dd6a35",
			DisplayName: "joonwi",
		},
	}
}

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/v3/emotes/{id}", func(w http.ResponseWriter, r *http.Request) {
		emoteID := chi.URLParam(r, "id")

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
