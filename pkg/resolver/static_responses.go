package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
)

var (
	NoLinkInfoFound  = []byte(`{"status":404,"message":"Could not fetch link info: No link info found"}`)
	InvalidURL       []byte
	ResponseTooLarge []byte
)

func InitializeStaticResponses(ctx context.Context, cfg config.APIConfig) {
	log := logger.FromContext(ctx)

	var err error
	var r *Response

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

func Errorf(format string, a ...interface{}) (*Response, time.Duration, error) {
	r := &Response{
		Status:  http.StatusInternalServerError,
		Message: CleanResponse(fmt.Sprintf(format, a...)),
	}

	return r, NoSpecialDur, nil
}
