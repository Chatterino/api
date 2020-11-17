package frankerfacez

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

func load(emoteID string, r *http.Request) (interface{}, error, time.Duration) {
	apiURL := fmt.Sprintf(emoteAPIURL, emoteID)
	thumbnailURL := fmt.Sprintf(thumbnailFormat, emoteID)

	// Create FrankerFaceZ API request
	resp, err := resolver.RequestGET(apiURL)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "ffz http request error " + resolver.CleanResponse(err.Error()),
		}, nil, cache.NoSpecialDur
	}
	defer resp.Body.Close()

	// Error out if the emote isn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return emoteNotFoundResponse, nil, cache.NoSpecialDur
	}

	// Read response into a string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "ffz http body read error " + resolver.CleanResponse(err.Error()),
		}, nil, cache.NoSpecialDur
	}

	// Parse response into a predefined JSON blob (see FrankerFaceZEmoteAPIResponse struct above)
	var temp struct {
		Emote FrankerFaceZEmoteAPIResponse `json:"emote"`
	}

	if err := json.Unmarshal(body, &temp); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "ffz api unmarshal error " + resolver.CleanResponse(err.Error()),
		}, nil, cache.NoSpecialDur
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
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "ffz template error " + resolver.CleanResponse(err.Error()),
		}, nil, cache.NoSpecialDur
	}

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnailURL,
	}, nil, cache.NoSpecialDur
}
