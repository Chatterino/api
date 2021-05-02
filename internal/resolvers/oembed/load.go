package oembed

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/dyatlov/go-oembed/oembed"
)

func load(fullURL string, r *http.Request) (interface{}, time.Duration, error) {
	item := oEmbed.FindItem(fullURL)

	data, err := item.FetchOembed(oembed.Options{URL: fullURL})

	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "oEmbed error: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	if data.Status >= 300 {
		fmt.Printf("[oEmbed] Skipping url %s because status code is %d\n", fullURL, data.Status)
		return &resolver.Response{
			Status:  data.Status,
			Message: fmt.Sprintf("oEmbed status code: %d", data.Status),
		}, cache.NoSpecialDur, nil
	}

	infoTooltipData := oEmbedData{data, fullURL}

	infoTooltipData.Title = humanize.Title(infoTooltipData.Title)
	infoTooltipData.Description = humanize.Description(infoTooltipData.Description)
	infoTooltipData.FullURL = fullURL

	// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
	var tooltip bytes.Buffer
	if err := oEmbedTemplate.Execute(&tooltip, infoTooltipData); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "oEmbed template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	resolverResponse := resolver.Response{
		Status:  200,
		Tooltip: url.PathEscape(tooltip.String()),
	}

	if infoTooltipData.Type == "photo" {
		resolverResponse.Thumbnail = infoTooltipData.URL
	}

	if infoTooltipData.ThumbnailURL != "" {

		// Some thumbnail URLs, like Streamable's returns // with no schema.
		if strings.HasPrefix(infoTooltipData.ThumbnailURL, "//") {
			infoTooltipData.ThumbnailURL = "https:" + infoTooltipData.ThumbnailURL
		}

		resolverResponse.Thumbnail = infoTooltipData.ThumbnailURL
	}

	return &resolverResponse, cache.NoSpecialDur, nil
}
