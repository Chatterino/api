package staticresponse

import (
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

type StaticResponse struct {
	payload     []byte
	statusCode  *int
	contentType *string

	cacheDuration time.Duration
}

func (s *StaticResponse) WithCacheDuration(cacheDuration time.Duration) *StaticResponse {
	s.cacheDuration = cacheDuration
	return s
}

func (s *StaticResponse) WithStatusCode(statusCode int) *StaticResponse {
	s.statusCode = &statusCode
	return s
}

func (s *StaticResponse) Return() ([]byte, *int, *string, time.Duration, error) {
	return s.payload, s.statusCode, s.contentType, s.cacheDuration, nil
}

func (s *StaticResponse) CacheError() (*cache.Response, error) {
	response := &cache.Response{
		Payload:     s.payload,
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
	}

	if s.statusCode != nil {
		response.StatusCode = *s.statusCode
	}
	if s.contentType != nil {
		response.ContentType = *s.contentType
	}

	return response, nil
}

func New(payload []byte) *StaticResponse {
	return &StaticResponse{
		payload:       payload,
		cacheDuration: resolver.NoSpecialDur,
	}
}
