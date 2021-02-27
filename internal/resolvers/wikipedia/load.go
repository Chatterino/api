package wikipedia

import (
	"net/http"
	"time"

	"log"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

type response struct {
	resolverResponse *resolver.Response
	err              error
}

func load(urlString string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[Wikipedia] GET", urlString)

	tooltipData, err := getPageInfo(urlString)

	if err != nil {
		log.Println("[Wikipedia] ERROR resolving URL", urlString, ":", err.Error())

		return response{
			resolverResponse: &resolver.Response{
				Status:  http.StatusOK,
				Tooltip: "Error getting Wikipedia API information for URL",
			},
			err: resolver.ErrDontHandle,
		}, cache.NoSpecialDur, nil
	}

	return buildTooltip(tooltipData)
}
