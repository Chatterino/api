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
	"github.com/Chatterino/api/pkg/utils"
)

func load(emoteHash string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[SevenTV] GET", emoteHash)

	queryMap := map[string]interface{}{
		"query": `
query fetchEmote($id: String!) {
	emote(id: $id) {
		visibility
		id
		name
		owner {
			id
			display_name
		}
	}
}`,
		"variables": map[string]string{
			"id": emoteHash,
		},
	}

	queryBytes, err := json.Marshal(queryMap)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "SevenTV API request error" + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// Execute SevenTV API request
	resp, err := resolver.RequestPOST(seventvAPIURL, string(queryBytes))
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
	if jsonResponse.Data.Emote == nil {
		return emoteNotFoundResponse, cache.NoSpecialDur, nil
	}

	// Determine type of the emote based on visibility flags
	visibility := jsonResponse.Data.Emote.Visibility
	var emoteType []string

	if utils.HasBits(visibility, EmoteVisibilityGlobal) {
		emoteType = append(emoteType, "Global")
	}

	if utils.HasBits(visibility, EmoteVisibilityPrivate) {
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
		Unlisted: utils.HasBits(visibility, EmoteVisibilityHidden),
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
		Thumbnail: utils.FormatThumbnailURL(baseURL, r, fmt.Sprintf(thumbnailFormat, emoteHash)),
		Link:      fmt.Sprintf("https://7tv.app/emotes/%s", emoteHash),
	}

	// Hide thumbnail for unlisted or hidden emotes pajaS
	if data.Unlisted {
		successTooltip.Thumbnail = ""
	}

	return successTooltip, cache.NoSpecialDur, nil
}
