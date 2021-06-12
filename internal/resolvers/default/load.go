package defaultresolver

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

func (dr *R) load(urlString string, r *http.Request) (interface{}, time.Duration, error) {
	requestUrl, err := url.Parse(urlString)
	if err != nil {
		return resolver.InvalidURL, cache.NoSpecialDur, nil
	}

	for _, m := range dr.customResolvers {
		if m.Check(requestUrl) {
			data, err := m.Run(requestUrl, r)

			if errors.Is(err, resolver.ErrDontHandle) {
				break
			}

			return data, cache.NoSpecialDur, err
		}
	}

	resp, err := resolver.RequestGET(requestUrl.String())
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

	// If the initial request URL is different from the response's apparent request URL,
	// we likely followed a redirect. Re-check the custom URL managers to see if the
	// page we were redirected to supports rich content. If not, continue with the
	// default tooltip.
	if requestUrl.String() != resp.Request.URL.String() {
		for _, m := range dr.customResolvers {
			if m.Check(resp.Request.URL) {
				data, err := m.Run(resp.Request.URL, r)

				if errors.Is(err, resolver.ErrDontHandle) {
					break
				}

				return data, cache.NoSpecialDur, err
			}
		}
	}

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

	limiter := &resolver.WriteLimiter{Limit: resolver.MaxContentLength}

	doc, err := goquery.NewDocumentFromReader(io.TeeReader(resp.Body, limiter))
	if err != nil {
		return utils.MarshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "html parser error (or download) " + resolver.CleanResponse(err.Error()),
		})
	}

	data := dr.defaultTooltipData(doc, r, resp)

	// Truncate title and description in case they're too long
	data.Truncate()

	// Sanitize potential html values
	data.Sanitize()

	var tooltip bytes.Buffer
	if err := defaultTooltip.Execute(&tooltip, data); err != nil {
		return utils.MarshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "template error " + resolver.CleanResponse(err.Error()),
		})
	}

	response := &resolver.Response{
		Status:    resp.StatusCode,
		Tooltip:   url.PathEscape(tooltip.String()),
		Link:      resp.Request.URL.String(),
		Thumbnail: data.ImageSrc,
	}

	if thumbnail.IsSupportedThumbnail(resp.Header.Get("content-type")) {
		response.Thumbnail = utils.FormatThumbnailURL(dr.baseURL, r, resp.Request.URL.String())
	}

	return utils.MarshalNoDur(response)
}
