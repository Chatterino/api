package oembed

import "github.com/dyatlov/go-oembed/oembed"

type oEmbedData struct {
	*oembed.Info
	FullURL string
}
