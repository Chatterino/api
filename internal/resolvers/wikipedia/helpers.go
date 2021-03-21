package wikipedia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

func getPageInfo(urlString string) (*wikipediaTooltipData, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	// Since the Wikipedia API is locale-dependant, we need the locale code.
	// For example, if you want to resolve a de.wikipedia.org link, you need
	// to ping the DE API endpoint.
	localeMatch := localeRegexp.FindStringSubmatch(u.Hostname())
	if len(localeMatch) != 2 {
		return nil, errLocaleMatch
	}

	localeCode := localeMatch[1]

	titleMatch := titleRegexp.FindStringSubmatch(u.Path)
	if len(titleMatch) != 2 {
		return nil, errTitleMatch
	}

	canonicalName := titleMatch[1]

	requestURL := fmt.Sprintf(endpointURL, localeCode, canonicalName)

	resp, err := resolver.RequestGET(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	var pageInfo *wikipediaAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&pageInfo); err != nil {
		return nil, err
	}

	// Transform API response into our tooltip model for Wikipedia links
	tooltipData := &wikipediaTooltipData{}

	sanitizedTitle := html.EscapeString(pageInfo.Titles.Display)
	tooltipData.Title = humanize.Title(sanitizedTitle)

	sanitizedExtract := html.EscapeString(pageInfo.Extract)
	tooltipData.Extract = humanize.Description(sanitizedExtract)

	if pageInfo.Description != nil {
		sanitizedDescription := html.EscapeString(*pageInfo.Description)
		tooltipData.Description = humanize.ShortDescription(sanitizedDescription)
	}

	if pageInfo.Thumbnail != nil {
		tooltipData.ThumbnailURL = pageInfo.Thumbnail.URL
	}

	return tooltipData, nil
}

func buildTooltip(pageInfo *wikipediaTooltipData) (response, time.Duration, error) {
	var tooltip bytes.Buffer

	if err := wikipediaTooltipTemplate.Execute(&tooltip, pageInfo); err != nil {
		return response{
			resolverResponse: &resolver.Response{
				Status:  http.StatusInternalServerError,
				Message: "Wikipedia template error: " + resolver.CleanResponse(err.Error()),
			}, err: nil,
		}, cache.NoSpecialDur, nil
	}

	return response{
		resolverResponse: &resolver.Response{
			Status:    http.StatusOK,
			Tooltip:   url.PathEscape(tooltip.String()),
			Thumbnail: pageInfo.ThumbnailURL,
		},
		err: nil,
	}, cache.NoSpecialDur, nil
}
