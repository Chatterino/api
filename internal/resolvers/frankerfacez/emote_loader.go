package frankerfacez

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

var (
	emoteNotFoundResponse = &resolver.Response{
		Status:  http.StatusNotFound,
		Message: "No FrankerFaceZ emote with this id found",
	}
)

/* Example JSON data generated from https://api.frankerfacez.com/v1/emote/131001 2020-11-18
{
  "emote": {
    "created_at": "2016-09-25T12:30:30.313Z",
    "css": null,
    "height": 21,
    "hidden": false,
    "id": 131001,
    "last_updated": "2016-09-25T14:25:01.408Z",
    "margins": null,
    "modifier": false,
    "name": "pajaE",
    "offset": null,
    "owner": {
      "_id": 63119,
      "display_name": "pajaSWA",
      "name": "pajaswa"
    },
    "public": true,
    "status": 1,
    "urls": {
      "1": "//cdn.frankerfacez.com/emote/131001/1",
      "2": "//cdn.frankerfacez.com/emote/131001/2",
      "4": "//cdn.frankerfacez.com/emote/131001/4"
    },
    "usage_count": 9,
    "width": 32
  }
}
*/

type EmoteAPIUser struct {
	DisplayName string `json:"display_name"`
	ID          int    `json:"_id"`
	Name        string `json:"name"`
}

type EmoteAPIResponse struct {
	Height    int16        `json:"height"`
	Modifier  bool         `json:"modifier"`
	Status    int          `json:"status"`
	Width     int16        `json:"width"`
	Hidden    bool         `json:"hidden"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"last_updated"`
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Public    bool         `json:"public"`
	Owner     EmoteAPIUser `json:"owner"`

	URLs struct {
		Size1 string `json:"1"`
		Size2 string `json:"2"`
		Size4 string `json:"4"`
	} `json:"urls"`
}

type TooltipData struct {
	Code     string
	Uploader string
}

type EmoteLoader struct {
	emoteAPIURL *url.URL
}

func (l *EmoteLoader) buildURL(emoteID string) string {
	relativeURL := &url.URL{
		Path: emoteID,
	}
	finalURL := l.emoteAPIURL.ResolveReference(relativeURL)

	return finalURL.String()
}

func (l *EmoteLoader) Load(ctx context.Context, emoteID string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)
	log.Debugw("Load FrankerFaceZ emote",
		"emoteID", emoteID,
	)
	apiURL := l.buildURL(emoteID)
	thumbnailURL := fmt.Sprintf(thumbnailFormat, emoteID)

	// Create FrankerFaceZ API request
	resp, err := resolver.RequestGET(ctx, apiURL)
	if err != nil {
		return resolver.Errorf("FrankerFaceZ HTTP request error: %s", err)
	}
	defer resp.Body.Close()

	// Error out if the emote isn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return emoteNotFoundResponse, cache.NoSpecialDur, nil
	}

	// Parse response into a predefined JSON blob (see EmoteAPIResponse struct above)
	var temp struct {
		Emote EmoteAPIResponse `json:"emote"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&temp); err != nil {
		return resolver.Errorf("FrankerFaceZ API unmarshal error: %s", err)
	}
	jsonResponse := temp.Emote

	// Build tooltip data from the API response
	data := TooltipData{
		Code:     jsonResponse.Name,
		Uploader: jsonResponse.Owner.DisplayName,
	}

	// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
	var tooltip bytes.Buffer
	if err := tmpl.Execute(&tooltip, data); err != nil {
		return resolver.Errorf("FrankerFaceZ template error: %s", err)
	}

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnailURL,
	}, cache.NoSpecialDur, nil
}

func NewEmoteLoader(emoteAPIURL *url.URL) *EmoteLoader {
	return &EmoteLoader{
		emoteAPIURL: emoteAPIURL,
	}
}
