package resolver

import (
	"net/url"
	"time"
)

maxContentSize, exists = utils.LookupEnv("MAX_CONTENT_SIZE")
if !exists {
	maxContentSize = 5
}

const (
	MaxContentLength = 1024 * 1024 * customMaxSize // IN MB
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
	Run   func(url *url.URL) ([]byte, error)
}

var NoSpecialDur time.Duration
