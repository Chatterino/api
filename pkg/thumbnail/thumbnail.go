package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"

	"github.com/Chatterino/api/pkg/config"
	"github.com/nfnt/resize"
)

var (
	supportedThumbnails = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

	cfg config.APIConfig
)

func InitializeConfig(passedCfg config.APIConfig) {
	cfg = passedCfg
}

// BuildStaticThumbnail is used when we fail to build an animated thumbnail using lilliput
func BuildStaticThumbnail(inputBuf []byte, resp *http.Response) ([]byte, error) {
	image, _, err := image.Decode(bytes.NewReader(inputBuf))
	if err != nil {
		return []byte{}, fmt.Errorf("could not decode image from url: %s", resp.Request.URL)
	}

	resized := resize.Thumbnail(cfg.MaxThumbnailSize, cfg.MaxThumbnailSize, image, resize.Bilinear)
	buffer := new(bytes.Buffer)
	if resp.Header.Get("content-type") == "image/png" {
		err = png.Encode(buffer, resized)
	} else if resp.Header.Get("content-type") == "image/gif" {
		err = gif.Encode(buffer, resized, nil)
	} else if resp.Header.Get("content-type") == "image/jpeg" {
		err = jpeg.Encode(buffer, resized, nil)
	}
	if err != nil {
		return []byte{}, fmt.Errorf("could not encode image from url: %s", resp.Request.URL)
	}

	return buffer.Bytes(), nil
}

func IsSupportedThumbnail(contentType string) bool {
	for _, supportedType := range supportedThumbnails {
		if contentType == supportedType {
			return true
		}
	}

	return false
}
