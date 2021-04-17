// +build windows

package defaultresolver

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"

	"github.com/nfnt/resize"
)

func buildThumbnailByteArray(resp *http.Response) ([]byte, error) {
	image, _, err := image.Decode(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("could not decode image from url: %s", resp.Request.URL)
	}

	resized := resize.Thumbnail(maxThumbnailSize, maxThumbnailSize, image, resize.Bilinear)
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
