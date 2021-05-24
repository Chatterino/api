package oembed

import "github.com/dyatlov/go-oembed/oembed"

type oEmbedData struct {
	*oembed.Info
	RequestedURL string
}

type facebookTokenResponse struct {
	AccessToken string `json:"access_token"`
}
