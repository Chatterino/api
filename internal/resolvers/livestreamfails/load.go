package livestreamfails

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

func load(clipID string, r *http.Request) (interface{}, time.Duration, error) {
	apiURL := fmt.Sprintf(livestreamfailsAPIURL, clipID)

	// Execute Livestreamfails API request
	resp, err := resolver.RequestGET(apiURL)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Livestreamfails http request error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}
	defer resp.Body.Close()

	// Error out if the clip isn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return noLivestreamfailsClipWithThisIDFound, cache.NoSpecialDur, nil
	}

	// Read response into a string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Livestreamfails http body read error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// Parse response into a predefined JSON blob (see Livestream struct above)
	var clipData LivestreamfailsAPIResponse
	if err := json.Unmarshal(body, &clipData); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Livestreamfails api unmarshal error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// Build tooltip data from the API response
	data := TooltipData{
		NSFW:         clipData.IsNSFW,
		Title:        clipData.Label,
		Category:     clipData.Category.Label,
		RedditScore:  humanize.Number(uint64(clipData.RedditScore)),
		Platform:     strings.Title(strings.ToLower(clipData.SourcePlatform)),
		StreamerName: clipData.Streamer.Label,
		CreationDate: humanize.CreationDate(clipData.CreatedAt),
	}

	// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
	var tooltip bytes.Buffer
	if err := livestreamfailsClipsTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Livestreamfails template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	resolverResponse := resolver.Response{
		Status:  200,
		Tooltip: url.PathEscape(tooltip.String()),
	}

	// Immediately return if the clip is marked NSFW.
	if clipData.IsNSFW {
		return &resolverResponse, cache.NoSpecialDur, nil
	}

	resolverResponse.Thumbnail = fmt.Sprintf(thumbnailFormat, clipData.ImageID)

	return &resolverResponse, cache.NoSpecialDur, nil
}
