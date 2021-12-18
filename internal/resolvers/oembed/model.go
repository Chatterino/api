package oembed

import "github.com/dyatlov/go-oembed/oembed"

type oEmbedData struct {
	*oembed.Info
	RequestedURL string
}

type facebookTokenResponse struct {
	AccessToken string `json:"access_token"`
}

const oEmbedTooltipString = `<div style="text-align: left;">
<b>{{.ProviderName}}{{ if .Title }} - {{.Title}}{{ end }}</b><hr>
{{ if .Description }}{{.Description}}{{ end }}
{{ if .AuthorName }}<br><b>Author:</b> {{.AuthorName}}{{ end }}
<br><b>URL:</b> {{.RequestedURL}}
</div>`
