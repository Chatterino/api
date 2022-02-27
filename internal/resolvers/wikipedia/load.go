package wikipedia

import (
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

func load(urlString string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log.Debugw("[Wikipedia] GET",
		"url", urlString,
	)

	tooltipData, err := getPageInfo(urlString)

	if err != nil {
		log.Debugw("[Wikipedia] Unable to get page info",
			"url", urlString,
			"error", err,
		)

		return nil, cache.NoSpecialDur, resolver.ErrDontHandle
	}

	return buildTooltip(tooltipData)
}
