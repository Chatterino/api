package betterttv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/resolver"
)

// Static responses
var (
	emoteNotFoundResponse = &resolver.Response{
		Status:  http.StatusNotFound,
		Message: "No BetterTTV emote with this hash found",
	}
)

// API structs

type EmoteAPIUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ProviderID  string `json:"providerId"`
}

type EmoteAPIResponse struct {
	ID             string       `json:"id"`
	Code           string       `json:"code"`
	ImageType      string       `json:"imageType"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
	Global         bool         `json:"global"`
	Live           bool         `json:"live"`
	Sharing        bool         `json:"sharing"`
	ApprovalStatus string       `json:"approvalStatus"`
	User           EmoteAPIUser `json:"user"`
}

// TODO: Should this live elsewhere?

type TooltipData struct {
	Code     string
	Type     string
	Uploader string
}

type EmoteLoader struct {
	emoteAPIURL string
}

func (l *EmoteLoader) Load(ctx context.Context, emoteHash string, r *http.Request) (*resolver.Response, time.Duration, error) {
	// TODO: Build URL smarter
	log := logger.FromContext(ctx)
	log.Debugw("Load BetterTTV emote",
		"emoteHash", emoteHash,
	)
	apiURL := fmt.Sprintf(l.emoteAPIURL, emoteHash)
	thumbnailURL := fmt.Sprintf(thumbnailFormat, emoteHash)

	// Create and execute BetterTTV API request
	resp, err := resolver.RequestGET(ctx, apiURL)
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
