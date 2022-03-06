package frankerfacez

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/go-chi/chi/v5"

	qt "github.com/frankban/quicktest"
)

func TestFoo(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	r := chi.NewRouter()
	r.Get("/v1/emote/{emoteID}", func(w http.ResponseWriter, r *http.Request) {
		emoteID := chi.URLParam(r, "emoteID")

		response := struct {
			Emote FrankerFaceZEmoteAPIResponse `json:"emote"`
		}{
			Emote: FrankerFaceZEmoteAPIResponse{
				Name:  emoteID,
				Owner: FrankerFaceZUser{"<b>B</b>", 123, "B"},
			},
		}

		b, _ := json.Marshal(&response)

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
	ts := httptest.NewServer(r)
	defer ts.Close()
	baseURL := ts.URL + "/v1/emote/%s"
	loader := &EmoteLoader{
		emoteAPIURL: baseURL,
	}

	response, _, err := loader.Load(ctx, "testemote", nil)

	c.Assert(err, qt.IsNil)
	c.Assert(response, qt.Not(qt.IsNil))

	c.Assert(response.Status, qt.Equals, 200)
	c.Assert(response.Thumbnail, qt.Equals, fmt.Sprintf(thumbnailFormat, "testemote"))

	const expectedTooltip = `<div style="text-align: left;">
<b>testemote</b><br>
<b>FrankerFaceZ Emote</b><br>
<b>By:</b> &lt;b&gt;B&lt;/b&gt;</div>`

	// TODO: check error
	cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
}
