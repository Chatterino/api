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
	statusCode := http.StatusOK
	contentType := "application/json"

	if s.statusCode != nil {
		statusCode = *s.statusCode
	}
	if s.contentType != nil {
		contentType = *s.contentType
	}

	return &cache.Response{
		Payload:     s.payload,
		StatusCode:  statusCode,
		ContentType: contentType,
	}, nil
}

func New(payload []byte) *StaticResponse {
	return &StaticResponse{
		payload:       payload,
		cacheDuration: resolver.NoSpecialDur,
	}
}
