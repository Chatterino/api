package defaultresolver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/staticresponse"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
)

type ThumbnailLoader struct {
	baseURL          string
	maxContentLength uint64
	enableLilliput   bool
}

func (l *ThumbnailLoader) Load(ctx context.Context, urlString string, r *http.Request) ([]byte, *int, *string, time.Duration, error) {
	log := logger.FromContext(ctx)

	url, err := url.Parse(urlString)
	if err != nil {
		return resolver.ReturnInvalidURL()
	}

	resp, err := resolver.RequestGET(ctx, url.String())
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			return resolver.InternalServerErrorf("Error loading thumbnail, could not resolve host %s", err.Error())
		}

		return resolver.InternalServerErrorf("Error loading thumbnail: %s", err.Error())
	}

	defer resp.Body.Close()

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		contentLengthBytes, err := strconv.Atoi(contentLength)
		if err != nil {
			r := &resolver.Response{
				Status:  http.StatusInternalServerError,
				Message: resolver.CleanResponse(fmt.Sprintf("Invalid content length: %s - %s", contentLength, err.Error())),
			}
			marshalledPayload, err := json.Marshal(r)
			if err != nil {
				panic(err)
			}

			return marshalledPayload, nil, nil, resolver.NoSpecialDur, nil
		}

		if uint64(contentLengthBytes) > l.maxContentLength {
			return resolver.FResponseTooLarge()
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		fmt.Println("Skipping url", resp.Request.URL, "because status code is", resp.StatusCode)
		return staticresponse.SNoThumbnailFound.Return()
	}

	contentType := resp.Header.Get("Content-Type")

	if !thumbnail.IsSupportedThumbnail(contentType) {
		return resolver.UnsupportedThumbnailType, nil, nil, cache.NoSpecialDur, nil
	}

	inputBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorw("Error reading body from request", "error", err)
		return resolver.ErrorBuildingThumbnail, nil, nil, cache.NoSpecialDur, nil
	}

	var image []byte
	// attempt building an animated image
	if l.enableLilliput {
		image, err = thumbnail.BuildAnimatedThumbnail(inputBuf, resp)
	}

	// fallback to static image if animated image building failed or is disabled
	if !l.enableLilliput || err != nil {
		if err != nil {
			log.Errorw("Error trying to build animated thumbnail, falling back to static thumbnail building",
				"error", err)
		}
		image, err = thumbnail.BuildStaticThumbnail(inputBuf, resp)
		if err != nil {
			log.Errorw("Error trying to build static thumbnail", "error", err)
			return resolver.InternalServerErrorf("Error building static thumbnail: %s", err.Error())
		}
	}

	return image, nil, &contentType, 10 * time.Minute, nil
}
