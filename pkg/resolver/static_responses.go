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
	UnsupportedThumbnailType = []byte(`{"status":415,"message":"Unsupported thumbnail type"}`)
	ErrorBuildingThumbnail   = []byte(`{"status":500,"message":"Error building thumbnail"}`)

	InvalidURLBytes = []byte(`{"status":400,"message":"Could not fetch link info: Invalid URL"}`)

	// Dynamically created based on config
	ResponseTooLarge []byte
)

func ReturnInvalidURL() ([]byte, *int, *string, time.Duration, error) {
	statusCode := http.StatusBadRequest
	contentType := "application/json"
	return InvalidURLBytes, &statusCode, &contentType, NoSpecialDur, nil
}

func WriteInvalidURL(w http.ResponseWriter) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	return w.Write(InvalidURLBytes)
}

func InitializeStaticResponses(ctx context.Context, cfg config.APIConfig) {
	log := logger.FromContext(ctx)

	var err error
	var r *Response

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

func WriteInternalServerErrorf(w http.ResponseWriter, format string, a ...interface{}) (int, error) {
	r := &Response{
		Status:  http.StatusInternalServerError,
		Message: CleanResponse(fmt.Sprintf(format, a...)),
	}

	marshalledPayload, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return w.Write(marshalledPayload)
}

func InternalServerErrorf(format string, a ...interface{}) ([]byte, *int, *string, time.Duration, error) {
	contentType := "application/json"

	r := &Response{
		Status:  http.StatusInternalServerError,
		Message: CleanResponse(fmt.Sprintf(format, a...)),
	}

	marshalledPayload, err := json.Marshal(r)
	if err != nil {
		return nil, nil, nil, NoSpecialDur, err
	}

	return marshalledPayload, nil, &contentType, NoSpecialDur, nil
}

func FResponseTooLarge() ([]byte, *int, *string, time.Duration, error) {
	statusCode := http.StatusInternalServerError
	contentType := "application/json"

	return ResponseTooLarge, &statusCode, &contentType, NoSpecialDur, nil
}
