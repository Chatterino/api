package wikipedia

import (
	"bytes"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

func buildTooltip(pageInfo *wikipediaTooltipData) (*resolver.Response, time.Duration, error) {
	var tooltip bytes.Buffer

	if err := wikipediaTooltipTemplate.Execute(&tooltip, pageInfo); err != nil {
		return resolver.Errorf("Wikipedia template error: %s", err.Error())
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: pageInfo.ThumbnailURL,
	}, cache.NoSpecialDur, nil
}
