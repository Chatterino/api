package defaultresolver

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
	"github.com/Chatterino/api/pkg/utils"
)

type ThumbnailLoader struct {
	baseURL          string
	maxContentLength uint64
	enableLilliput   bool
}

func (l *ThumbnailLoader) Load(ctx context.Context, urlString string, r *http.Request) ([]byte, time.Duration, error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return resolver.InvalidURL, cache.NoSpecialDur, nil
	}

	resp, err := resolver.RequestGET(ctx, url.String())
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
		if uint64(contentLengthBytes) > l.maxContentLength {
			return resolver.ResponseTooLarge, cache.NoSpecialDur, nil
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		fmt.Println("Skipping url", resp.Request.URL, "because status code is", resp.StatusCode)
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	if !thumbnail.IsSupportedThumbnail(resp.Header.Get("content-type")) {
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	inputBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading body from request:", err)
		return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
	}

	var image []byte
	// attempt building an animated image
	if l.enableLilliput {
		image, err = thumbnail.BuildAnimatedThumbnail(inputBuf, resp)
	}

	// fallback to static image if animated image building failed or is disabled
	if !l.enableLilliput || err != nil {
		if err != nil {
			log.Println("Error trying to build animated thumbnail:", err.Error(), "falling back to static thumbnail building")
		}
		image, err = thumbnail.BuildStaticThumbnail(inputBuf, resp)
		if err != nil {
			log.Println("Error trying to build static thumbnail:", err.Error())
			return resolver.NoLinkInfoFound, cache.NoSpecialDur, nil
		}
	}

	return image, 10 * time.Minute, nil
}
