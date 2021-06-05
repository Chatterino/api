package seventv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

func load(emoteHash string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[SevenTV] GET", emoteHash)

	// Execute SevenTV API request
	resp, err := resolver.RequestPOST(seventvAPIURL, fmt.Sprintf(gqlQueryEmotes, emoteHash))
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "SevenTV API request error" + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}
	defer resp.Body.Close()

	// Read response into a string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "SevenTV API http body read error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// Error out if the emote wasn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return emoteNotFoundResponse, cache.NoSpecialDur, nil
	}

	var jsonResponse EmoteAPIResponse
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "SevenTV API unmarshal error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// API returns Data.Emote as null if the emote wasn't found
	fmt.Println(jsonResponse.Data.Emote)
	if jsonResponse.Data.Emote == nil {
		return emoteNotFoundResponse, cache.NoSpecialDur, nil
	}

	// Build tooltip data from the API response
	data := TooltipData{
		Code:     jsonResponse.Data.Emote.Name,
		Type:     "Channel",
		Uploader: jsonResponse.Data.Emote.Owner.DisplayName,
	}

	// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
	var tooltip bytes.Buffer
	if err := seventvEmoteTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "SevenTV emote template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// Success
	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: fmt.Sprintf(thumbnailFormat, emoteHash),
		Link:      fmt.Sprintf("https://7tv.app/emotes/%s", emoteHash),
	}, cache.NoSpecialDur, nil
}
