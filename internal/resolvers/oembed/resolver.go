package oembed

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/dyatlov/go-oembed/oembed"
)

const (
	oEmbedTooltipString = `<div style="text-align: left;">
<b>{{.ProviderName}}{{ if .Title }} - {{.Title}}{{ end }}</b><hr>
{{ if .Description }}{{.Description}}{{ end }}
{{ if .AuthorName }}<br><b>Author:</b> {{.AuthorName}}{{ end }}
<br><b>URL:</b> {{.RequestedURL}}
</div>`
)

var (
	oEmbedTemplate = template.Must(template.New("oEmbedTemplateTooltip").Parse(oEmbedTooltipString))

	oEmbedCache = cache.New("oEmbed", load, 1*time.Hour)

	oEmbed = oembed.NewOembed()
)

func New() (resolvers []resolver.CustomURLManager) {
	providersPath := "./providers.json"

	if providersPathEnv, exists := utils.LookupEnv("OEMBED_PROVIDERS_PATH"); exists {
		log.Println("[oEmbed] Overriding path of providers.json to", providersPathEnv)
		providersPath = providersPathEnv
	}

	data, err := ioutil.ReadFile(providersPath)

	if err != nil {
		log.Println("[oEmbed] No providers.json file found, won't do oEmbed parsing")
		return
	}

	if facebookAppID, facebookAppSecret, exists := loadFacebookCredentials(); exists {
		if err := initFacebookAppAccessToken(facebookAppID, facebookAppSecret); err != nil {
			log.Println("[oEmbed] error loading facebook app access token", err)
		} else {
			log.Println("[oEmbed] Extra rich info loading enabled for Instagram and Facebook")
		}
	}

	oEmbed.ParseProviders(bytes.NewReader(data))

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return oEmbed.FindItem(url.String()) != nil
		},
		Run: func(url *url.URL) ([]byte, error) {
			apiResponse := oEmbedCache.Get(url.String(), nil)
			return json.Marshal(apiResponse)
		},
	})

	return
}
