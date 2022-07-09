package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type InnerLoader interface {
	Load(ctx context.Context, key string, r *http.Request) (*Response, time.Duration, error)
}

type ResponseMarshaller struct {
	innerLoader InnerLoader
}

func (m *ResponseMarshaller) Load(ctx context.Context, s string, r *http.Request) ([]byte, *int, *string, time.Duration, error) {
	value, specialDur, err := m.innerLoader.Load(ctx, s, r)
	if err != nil {
		return nil, nil, nil, specialDur, err
	}

	if value == nil {
		return nil, nil, nil, specialDur, errors.New("inner load value must not be nil when error is nil")
	}

	valueBytes, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		return nil, nil, nil, specialDur, marshalErr
	}

	return valueBytes, nil, nil, specialDur, nil
}

func NewResponseMarshaller(innerLoader InnerLoader) *ResponseMarshaller {
	if innerLoader == nil {
		log.Fatalf("NewResponseMarshaller called with a nil innerLoader")
	}

	m := &ResponseMarshaller{
		innerLoader: innerLoader,
	}

	return m
}
