// TODO: this should potentially be split out into its own file
package defaultresolver

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/nfnt/resize"
)

var (
	supportedThumbnails = []string{"image/jpeg", "image/png", "image/gif"}
)

const (
	// max width or height the thumbnail will be resized to
	maxThumbnailSize = 300
)

func doThumbnailRequest(urlString string, r *http.Request) (interface{}, time.Duration, error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return resolver.InvalidURL, cache.NoSpecialDur, nil
	}

	resp, err := resolver.RequestGET(url.String())
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
		}

		return utils.MarshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: resolver.CleanResponse(err.Error()),
		})
	}

	defer resp.Body.Close()

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		contentLengthBytes, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, cache.NoSpecialDur, err
		}
		if contentLengthBytes > resolver.MaxContentLength {
			return resolver.ResponseTooLarge, cache.NoSpecialDur, nil
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		fmt.Println("Skipping url", resp.Request.URL, "because status code is", resp.StatusCode)
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	if !isSupportedThumbnail(resp.Header.Get("content-type")) {
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	image, err := buildThumbnailByteArray(resp)
	if err != nil {
		log.Println(err.Error())
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	return image, 10 * time.Minute, nil
}

func isSupportedThumbnail(contentType string) bool {
	for _, supportedType := range supportedThumbnails {
		if contentType == supportedType {
			return true
		}
	}

	return false
}

func thumbnail(w http.ResponseWriter, r *http.Request) {
	url, err := utils.UnescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(resolver.InvalidURL)
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	response := thumbnailCache.Get(url, r)

	_, err = w.Write(response.([]byte))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

var (
	thumbnailCache = cache.New("thumbnail", doThumbnailRequest, 10*time.Minute)
)

func InitializeThumbnail(router *chi.Mux) {
	router.Get("/thumbnail/{url}", thumbnail)
}

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
