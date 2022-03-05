package wikipedia

import (
	"context"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

type ArticleLoader struct {
	emoteAPIURL string
}

func (l *ArticleLoader) Load(ctx context.Context, urlString string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Wikipedia] GET",
		"url", urlString,
	)

	tooltipData, err := getPageInfo(ctx, urlString)

	if err != nil {
		log.Debugw("[Wikipedia] Unable to get page info",
			"url", urlString,
			"error", err,
		)

		return nil, cache.NoSpecialDur, resolver.ErrDontHandle
	}

	return buildTooltip(tooltipData)
}
