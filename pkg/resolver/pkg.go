package resolver

import (
	"net/url"
	"strconv"
	"time"

	"github.com/Chatterino/api/pkg/utils"
)

func getMaxContentLength() int {
	maxContentLengthStr, exists := utils.LookupEnv("MAX_CONTENT_LENGTH")
	maxContentLength := 5
	if exists {
		if p, err := strconv.ParseFloat(maxContentLengthStr, 32); err != nil {
			maxContentLength := float64(5)
		}
	}
	return maxContentLength * 1024 * 1024
}

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
