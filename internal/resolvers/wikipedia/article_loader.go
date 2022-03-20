package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

type ArticleLoader struct {
	// the apiURL format must consist of 2 %s, first being region second being article
	apiURL string
}

func (l *ArticleLoader) Load(ctx context.Context, unused string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	// Since the Wikipedia API is locale-dependant, we need the locale code.
	// For example, if you want to resolve a de.wikipedia.org link, you need
	// to ping the DE API endpoint.
	// If no locale is specified in the given URL, we will assume it's the english wiki article
	localeCode, articleID, err := articleValuesFromContext(ctx)
	if err != nil {
		return nil, resolver.NoSpecialDur, err
	}

	log.Debugw("[Wikipedia] GET",
		"localeCode", localeCode,
		"articleID", articleID,
	)

	requestURL := fmt.Sprintf(l.apiURL, localeCode, articleID)

	resp, err := resolver.RequestGET(ctx, requestURL)
	if err != nil {
		return nil, resolver.NoSpecialDur, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &resolver.Response{
			Status:  http.StatusNotFound,
			Message: "No Wikipedia article found",
		}, resolver.NoSpecialDur, nil
		// return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	var pageInfo *wikipediaAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&pageInfo); err != nil {
		return nil, resolver.NoSpecialDur, err
	}

	// Transform API response into our tooltip model for Wikipedia links
	tooltipData := &wikipediaTooltipData{}

	sanitizedTitle := pageInfo.Titles.Normalized
	tooltipData.Title = humanize.Title(sanitizedTitle)

	sanitizedExtract := pageInfo.Extract
	tooltipData.Extract = humanize.Description(sanitizedExtract)

	if pageInfo.Description != nil {
		sanitizedDescription := *pageInfo.Description
		tooltipData.Description = humanize.ShortDescription(sanitizedDescription)
	}

	if pageInfo.Thumbnail != nil {
		tooltipData.ThumbnailURL = pageInfo.Thumbnail.URL
	}

	return buildTooltip(tooltipData)
}
