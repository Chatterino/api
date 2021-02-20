package betterttv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/resolver"
)

func load(emoteHash string, r *http.Request) (interface{}, time.Duration, error) {
	apiURL := fmt.Sprintf(emoteAPIURL, emoteHash)
	thumbnailURL := fmt.Sprintf(thumbnailFormat, emoteHash)

	// Create and execute BetterTTV API request
	resp, err := resolver.RequestGET(apiURL)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "betterttv http request error " + resolver.CleanResponse(err.Error()),
		}, resolver.NoSpecialDur, nil
	}
	defer resp.Body.Close()

	// Error out if the emote isn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return emoteNotFoundResponse, resolver.NoSpecialDur, nil
	}

	// Read response into a string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "betterttv http body read error " + resolver.CleanResponse(err.Error()),
		}, resolver.NoSpecialDur, nil
	}

	// Parse response into a predefined JSON blob (see EmoteAPIResponse struct in model.go)
	var jsonResponse EmoteAPIResponse
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "betterttv api unmarshal error " + resolver.CleanResponse(err.Error()),
		}, resolver.NoSpecialDur, nil
	}

	// Build tooltip data from the API response
	data := TooltipData{
		Code:     jsonResponse.Code,
		Type:     "Shared",
		Uploader: jsonResponse.User.DisplayName,
	}

	if jsonResponse.Global {
		data.Type = "Global"
	}

	// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
	var tooltip bytes.Buffer
	if err := tmpl.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "youtube template error " + resolver.CleanResponse(err.Error()),
		}, resolver.NoSpecialDur, nil
	}

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnailURL,
	}, resolver.NoSpecialDur, nil
}
