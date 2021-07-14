package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/nfnt/resize"
)

var (
	supportedThumbnails = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

	cfg config.APIConfig
)

func InitializeConfig(passedCfg config.APIConfig) {
	cfg = passedCfg
}

// buildStaticThumbnailByteArray is used when we fail to build an animated thumbnail using lilliput
func buildStaticThumbnailByteArray(inputBuf []byte, resp *http.Response) ([]byte, error) {
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

func DoThumbnailRequest(urlString string, r *http.Request) (interface{}, time.Duration, error) {
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
		if uint64(contentLengthBytes) > cfg.MaxContentLength {
			return resolver.ResponseTooLarge, cache.NoSpecialDur, nil
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		fmt.Println("Skipping url", resp.Request.URL, "because status code is", resp.StatusCode)
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	if !IsSupportedThumbnail(resp.Header.Get("content-type")) {
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	inputBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading body from request:", err)
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	var image []byte
	// attempt building an animated image
	if cfg.EnableLilliput {
		image, err = buildThumbnailByteArray(inputBuf, resp)
	}

	// fallback to static image if animated image building failed or is disabled
	if !cfg.EnableLilliput || err != nil {
		if err != nil {
			log.Println("Error trying to build animated thumbnail:", err.Error(), "falling back to static thumbnail building")
		}
		image, err = buildStaticThumbnailByteArray(inputBuf, resp)
		if err != nil {
			log.Println("Error trying to build static thumbnail:", err.Error())
			return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
		}
	}

	return image, 10 * time.Minute, nil
}

func IsSupportedThumbnail(contentType string) bool {
	for _, supportedType := range supportedThumbnails {
		if contentType == supportedType {
			return true
		}
	}

	return false
}
