package livestreamfails

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type ClipAPICategory struct {
	Label string `json:"label"`
}

type ClipAPIStreamer struct {
	Label string `json:"label"`
}

type ClipAPIResponse struct {
	Category       ClipAPICategory `json:"category"`
	CreatedAt      time.Time       `json:"createdAt"`
	ImageID        string          `json:"imageId"`
	IsNSFW         bool            `json:"isNSFW"`
	Label          string          `json:"label"`
	RedditScore    int             `json:"redditScore"`
	SourcePlatform string          `json:"sourcePlatform"`
	Streamer       ClipAPIStreamer `json:"streamer"`
}

type ClipLoader struct {
	apiURLFormat string
}

func (l *ClipLoader) Load(ctx context.Context, clipID string, r *http.Request) (*resolver.Response, time.Duration, error) {
	apiURL := fmt.Sprintf(l.apiURLFormat, clipID)

	// Execute Livestreamfails API request
	resp, err := resolver.RequestGET(ctx, apiURL)
	if err != nil {
		return resolver.Errorf("Livestreamfails HTTP request error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return noLivestreamfailsClipWithThisIDFound, cache.NoSpecialDur, nil
	}

	// Error out if the clip isn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return resolver.Errorf("Livestreamfails unhandled HTTP status code: %d", resp.StatusCode)
	}

	// Parse response into a predefined JSON format
	var clipData ClipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&clipData); err != nil {
		return resolver.Errorf("Livestreamfails API response decode error: %s", err)
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
		return resolver.Errorf("Livestreamfails template error: %s", err)
	}

	resolverResponse := resolver.Response{
		Status:  http.StatusOK,
		Tooltip: url.PathEscape(tooltip.String()),
	}

	// Immediately return if the clip is marked NSFW.
	if clipData.IsNSFW {
		return &resolverResponse, cache.NoSpecialDur, nil
	}

	resolverResponse.Thumbnail = fmt.Sprintf(thumbnailFormat, clipData.ImageID)

	return &resolverResponse, cache.NoSpecialDur, nil
}
