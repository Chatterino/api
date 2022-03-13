package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

type ArticleLoader struct {
	endpointURL string
}

func (l *ArticleLoader) getPageInfo(ctx context.Context, urlString string) (*wikipediaTooltipData, error) {
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

	requestURL := fmt.Sprintf(l.endpointURL, localeCode, canonicalName)

	resp, err := resolver.RequestGET(ctx, requestURL)
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

	return tooltipData, nil
}

func (l *ArticleLoader) Load(ctx context.Context, urlString string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Wikipedia] GET",
		"url", urlString,
	)

	tooltipData, err := l.getPageInfo(ctx, urlString)

	if err != nil {
		log.Debugw("[Wikipedia] Unable to get page info",
			"url", urlString,
			"error", err,
		)

		return nil, cache.NoSpecialDur, resolver.ErrDontHandle
	}

	return buildTooltip(tooltipData)
}
