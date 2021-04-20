//go:generate mockgen -destination ../../mocks/mock_imgurClient.go -package=mocks . ImgurClient
// The above comment will make it so that when `go generate` is called, the command after go:generate is called in this files PWD.
// The mockgen command itself generates a mock for the ImgurClient interface in this file and stores it in the internal/mocks/ package

package imgur

import (
	"html/template"
	"os"
	"time"

	"log"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/koffeinsource/go-imgur"
)

type ImgurClient interface {
	GetInfoFromURL(urlString string) (*imgur.GenericInfo, int, error)
}

var (
	// max size of an image before we use a small thumbnail of it
	maxRawImageSize = 50 * 1024

	imageTooltipTemplate = template.Must(template.New("imageTooltipTemplate").Parse(imageTooltip))

	imgurCache = cache.New("imgur", load, 1*time.Hour)

	apiClient ImgurClient
)

func New() (resolvers []resolver.CustomURLManager) {
	var clientID string
	var exists bool

	if clientID, exists = os.LookupEnv("CHATTERINO_API_IMGUR_CLIENT_ID"); !exists {
		log.Println("No CHATTERINO_API_IMGUR_CLIENT_ID specified, won't do special responses for imgur")
		return
	}

	apiClient = &imgur.Client{
		HTTPClient:    resolver.HTTPClient(),
		Log:           &NullLogger{},
		ImgurClientID: clientID,
		RapidAPIKEY:   "",
	}

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
