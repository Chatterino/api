package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
)

var (
	NoLinkInfoFound  []byte
	InvalidURL       []byte
	ResponseTooLarge []byte
)

func InitializeStaticResponses(ctx context.Context, cfg config.APIConfig) {
	log := logger.FromContext(ctx)

	var err error
	r := &Response{
		Status:  404,
		Message: "Could not fetch link info: No link info found",
	}

	NoLinkInfoFound, err = json.Marshal(r)
	if err != nil {
		log.Fatalw("Error marshalling prebuilt response",
			"error", err,
		)
	}

	r = &Response{
		Status:  500,
		Message: "Could not fetch link info: Invalid URL",
	}
	InvalidURL, err = json.Marshal(r)
	if err != nil {
		log.Fatalw("Error marshalling prebuilt response",
			"error", err,
		)
	}

	r = &Response{
		Status:  http.StatusInternalServerError,
		Message: fmt.Sprintf("Could not fetch link info: Response too large (>%dMB)", cfg.MaxContentLength/1024/1024),
	}
	ResponseTooLarge, err = json.Marshal(r)
	if err != nil {
		log.Fatalw("Error marshalling prebuilt response",
			"error", err,
		)
	}
}
