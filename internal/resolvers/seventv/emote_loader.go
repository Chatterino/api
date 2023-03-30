package seventv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type EmoteLoader struct {
	apiURL  string
	baseURL string
}

func (l *EmoteLoader) Load(ctx context.Context, emoteHash string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[SevenTV] Get emote",
		"emoteHash", emoteHash,
	)

	// Execute SevenTV API request
	resp, err := resolver.RequestGET(ctx, fmt.Sprintf("%s/%s", l.apiURL, emoteHash))
	if err != nil {
		return resolver.Errorf("7TV API request error: %s", err)
	}
	defer resp.Body.Close()

	// Error out if the emote wasn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return emoteNotFoundResponse, cache.NoSpecialDur, nil
	}

	var jsonResponse EmoteModel
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return resolver.Errorf("7TV API response decode error: %s", err)
	}

	var emoteType string
	if utils.HasBits(int32(jsonResponse.Flags), int32(EmoteFlagsPrivate)) {
		emoteType = "Private"
	} else {
		emoteType = "Shared"
	}

	// Build tooltip data from the API response
	data := TooltipData{
		Code:     jsonResponse.Name,
		Type:     emoteType,
		Uploader: jsonResponse.Owner.DisplayName,
		Unlisted: !jsonResponse.Listed,
	}

	// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
	var tooltip bytes.Buffer
	if err := seventvEmoteTemplate.Execute(&tooltip, data); err != nil {
		return resolver.Errorf("7TV emote template error: %s", err)
	}

	var bestFile *ImageFile
	var bestWidth int32
	for _, file := range jsonResponse.Host.Files {
		if file.Format == ImageFormatWEBP && file.Width > bestWidth {
			bestFile = &file
			bestWidth = file.Width
		}
	}
	var thumbnail string
	// Hide thumbnail for unlisted or hidden emotes pajaS
	if !data.Unlisted && bestFile != nil {
		if strings.HasPrefix(jsonResponse.Host.URL, "//") {
			thumbnail = fmt.Sprintf("https:%s/%s", jsonResponse.Host.URL, bestFile.Name)
		} else {
			thumbnail = fmt.Sprintf("%s/%s", jsonResponse.Host.URL, bestFile.Name)
		}
		thumbnail = utils.FormatThumbnailURL(l.baseURL, r, thumbnail)
	}

	// Success
	successTooltip := &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnail,
		Link:      fmt.Sprintf("https://7tv.app/emotes/%s", emoteHash),
	}

	return successTooltip, cache.NoSpecialDur, nil
}

func NewEmoteLoader(cfg config.APIConfig, apiURL *url.URL) *EmoteLoader {
	return &EmoteLoader{
		apiURL:  apiURL.String(),
		baseURL: cfg.BaseURL,
	}
}
