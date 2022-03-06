package livestreamfails

import (
	"bytes"
	"context"
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

type TooltipData struct {
	NSFW         bool
	Title        string
	Category     string
	RedditScore  string
	Platform     string
	StreamerName string
	CreationDate string
}

type Resize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Output struct {
	Format string `json:"format"`
}

type LivestreamFailsThumbnailRequest struct {
	Input  string `json:"input"`
	Resize Resize `json:"resize"`
	Output Output `json:"output"`
}

type LivestreamfailsAPIResponse struct {
	Category struct {
		Label string `json:"label"`
	} `json:"category"`
	CreatedAt      time.Time `json:"createdAt"`
	ImageID        string    `json:"imageId"`
	IsNSFW         bool      `json:"isNSFW"`
	Label          string    `json:"label"`
	RedditScore    int       `json:"redditScore"`
	SourcePlatform string    `json:"sourcePlatform"`
	Streamer       struct {
		Label string `json:"label"`
	} `json:"streamer"`
}

type ClipLoader struct {
}

func (l *ClipLoader) Load(ctx context.Context, clipID string, r *http.Request) (*resolver.Response, time.Duration, error) {
	apiURL := fmt.Sprintf(livestreamfailsAPIURL, clipID)

	// Execute Livestreamfails API request
	resp, err := resolver.RequestGET(ctx, apiURL)
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
