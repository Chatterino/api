package defaultresolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/staticresponse"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

type LinkLoader struct {
	baseURL          string
	customResolvers  []resolver.Resolver
	maxContentLength uint64
}

func (l *LinkLoader) defaultTooltipData(doc *goquery.Document, r *http.Request, resp *http.Response) tooltipData {
	data := tooltipMetaFields(l.baseURL, doc, r, resp, tooltipData{
		URL: resolver.CleanResponse(resp.Request.URL.String()),
	})

	if data.Title == "" {
		data.Title = doc.Find("title").First().Text()
	}

	return data
}

func (l *LinkLoader) Load(ctx context.Context, urlString string, r *http.Request) ([]byte, *int, *string, time.Duration, error) {
	requestUrl, err := url.Parse(urlString)
	if err != nil {
		return resolver.ReturnInvalidURL()
	}

	resp, err := resolver.RequestGET(ctx, requestUrl.String())
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			return staticresponse.SNoLinkInfoFound.
				Return()
		}

		return staticresponse.InternalServerErrorf("Error loading link: %s", err.Error()).
			WithStatusCode(http.StatusInternalServerError).
			Return()
	}

	defer resp.Body.Close()

	// If the initial request URL is different from the response's apparent request URL,
	// we likely followed a redirect. Re-check the custom URL managers to see if the
	// page we were redirected to supports rich content. If not, continue with the
	// default tooltip.
	if requestUrl.String() != resp.Request.URL.String() {
		for _, m := range l.customResolvers {
			if ctx, result := m.Check(ctx, resp.Request.URL); result {
				data, err := m.Run(ctx, resp.Request.URL, r)

				if errors.Is(err, resolver.ErrDontHandle) {
					break
				}

				return data.Payload, &data.StatusCode, &data.ContentType, cache.NoSpecialDur, err
			}
		}
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		contentLengthBytes, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, nil, nil, cache.NoSpecialDur, err
		}
		if uint64(contentLengthBytes) > l.maxContentLength {
			return resolver.ResponseTooLarge, nil, nil, cache.NoSpecialDur, nil
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		fmt.Println("Skipping url", resp.Request.URL, "because status code is", resp.StatusCode)
		return staticresponse.SNoLinkInfoFound.Return()
	}

	limiter := &resolver.WriteLimiter{Limit: l.maxContentLength}

	doc, err := goquery.NewDocumentFromReader(io.TeeReader(resp.Body, limiter))
	if err != nil {
		return utils.MarshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "html parser error (or download) " + resolver.CleanResponse(err.Error()),
		})
	}

	data := l.defaultTooltipData(doc, r, resp)

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
		response.Thumbnail = utils.FormatThumbnailURL(l.baseURL, r, resp.Request.URL.String())
	}

	return utils.MarshalNoDur(response)
}
