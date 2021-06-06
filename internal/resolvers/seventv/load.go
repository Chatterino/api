package seventv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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

	// Determine type of the emote based on visibility flags
	visibility := jsonResponse.Data.Emote.Visibility
	var emoteType []string

	if visibility&EmoteVisibilityGlobal != 0 {
		emoteType = append(emoteType, "Global")
	}

	if visibility&EmoteVisibilityPrivate != 0 {
		emoteType = append(emoteType, "Private")
	}

	// Default to Shared emote
	if len(emoteType) == 0 {
		emoteType = append(emoteType, "Shared")
	}

	// Build tooltip data from the API response
	data := TooltipData{
		Code:     jsonResponse.Data.Emote.Name,
		Type:     strings.Join(emoteType, " "),
		Uploader: jsonResponse.Data.Emote.Owner.DisplayName,
		Unlisted: visibility&EmoteVisibilityHidden != 0,
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
	successTooltip := &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: fmt.Sprintf(thumbnailFormat, emoteHash),
		Link:      fmt.Sprintf("https://7tv.app/emotes/%s", emoteHash),
	}

	// Hide thumbnail for unlisted or hidden emotes pajaS
	if data.Unlisted {
		successTooltip.Thumbnail = ""
	}

	return successTooltip, cache.NoSpecialDur, nil
}
