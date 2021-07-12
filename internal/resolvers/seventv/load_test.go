package seventv

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Chatterino/api/pkg/resolver"
	"github.com/go-chi/chi/v5"

	qt "github.com/frankban/quicktest"
)

var (
	emotes = map[string]EmoteAPIResponse{}
)

func init() {
	emotes["Pajawalk"] = EmoteAPIResponse{
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

	emotes["Hidden"] = EmoteAPIResponse{
		Data: EmoteAPIResponseData{
			Emote: &EmoteAPIEmote{
				ID:         "604281c81ae70f000d47ffd9",
				Name:       "Hidden",
				Visibility: EmoteVisibilityPrivate | EmoteVisibilityHidden,
				Owner: EmoteAPIUser{
					ID:          "603d10e496832ffa787ca53c",
					DisplayName: "durado_",
				},
			},
		},
	}
}

func TestFoo(t *testing.T) {
	c := qt.New(t)

	r := chi.NewRouter()
	r.Post("/v2/gql", func(w http.ResponseWriter, r *http.Request) {
		type gqlQuery struct {
			Query     string            `json:"string"`
			Variables map[string]string `json:"variables"`
		}

		var q gqlQuery

		xd, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(xd, &q)
		if err != nil {
			panic(err)
		}
		if response, ok := emotes[q.Variables["id"]]; ok {
			b, _ := json.Marshal(&response)

			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
			return
		}

		// TODO: return 404
	})
	ts := httptest.NewServer(r)
	defer ts.Close()
	seventvAPIURL = ts.URL + "/v2/gql"

	type tTest struct {
		emoteHash       string
		expectedTooltip string
	}

	tests := []tTest{
		{
			emoteHash: "Pajawalk",
			expectedTooltip: `<div style="text-align: left;">
<b>Pajawalk</b><br>
<b>Private SevenTV Emote</b><br>
<b>By:</b> durado_
</div>`,
		},
		{
			emoteHash: "Hidden",
			expectedTooltip: `<div style="text-align: left;">
<b>Hidden</b><br>
<b>Private SevenTV Emote</b><br>
<b>By:</b> durado_
<li><b><span style="color: red;">UNLISTED</span></b></li>
</div>`,
		},
		// TODO: Global emote
		// TODO: Private emote
		// TODO: Combined emote types
		// TODO: Default emote type (shared)
		// TODO: Thumbnails
		// TODO: emote not found (404)
	}

	request, _ := http.NewRequest(http.MethodPost, "https://7tv.app/test", nil)

	for _, test := range tests {
		iret, _, err := load(test.emoteHash, request)

		c.Assert(err, qt.IsNil)
		c.Assert(iret, qt.Not(qt.IsNil))

		response := iret.(*resolver.Response)

		c.Assert(response, qt.Not(qt.IsNil))

		c.Assert(response.Status, qt.Equals, 200)

		// TODO: check thumbnail
		// c.Assert(response.Thumbnail, qt.Equals, fmt.Sprintf(thumbnailFormat, test.emoteHash))

		// TODO: check error
		cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
		c.Assert(unescapeErr, qt.IsNil)

		c.Assert(cleanTooltip, qt.Equals, test.expectedTooltip)
	}
}
