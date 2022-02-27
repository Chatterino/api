package resolver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`

	Thumbnail string `json:"thumbnail,omitempty"`
	Tooltip   string `json:"tooltip,omitempty"`
	Link      string `json:"link,omitempty"`

	// Flag in the BTTV API to.. maybe signify that the link will download something? idk
	// Download *bool  `json:"download,omitempty"`
}

type CustomURLManager struct {
	Check func(url *url.URL) bool
	Run   func(url *url.URL, r *http.Request) ([]byte, error)
}

var NoSpecialDur time.Duration

// MarshalResponse can take a loader function that returns a Response struct and ensure it gets marshalled correctly
func MarshalResponse(innerLoad func(s string, r *http.Request) (*Response, time.Duration, error)) func(s string, r *http.Request) ([]byte, time.Duration, error) {
	return func(s string, r *http.Request) ([]byte, time.Duration, error) {
		value, specialDur, err := innerLoad(s, r)
		if err != nil {
			return nil, specialDur, err
		}

		if value == nil {
			return nil, specialDur, errors.New("inner load value must not be nil when error is nil")
		}

		valueBytes, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			return nil, specialDur, marshalErr
		}

		return valueBytes, specialDur, nil
	}
}
