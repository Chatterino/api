package defaultresolver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/resolvers/twitter"
	"github.com/Chatterino/api/internal/staticresponse"
	"github.com/Chatterino/api/internal/version"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

type LinkLoader struct {
	baseURL              string
	customResolvers      []resolver.Resolver
	contentTypeResolvers []ContentTypeResolver
	maxContentLength     uint64
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
	log := logger.FromContext(ctx)

	requestUrl, err := url.Parse(urlString)
	if err != nil {
		return resolver.ReturnInvalidURL()
	}

	extraHeaders := make(map[string]string)
	cacheDur := cache.NoSpecialDur
	ctx, isTwitterRequest := twitter.Check(ctx, requestUrl)
	if isTwitterRequest {
		extraHeaders["User-Agent"] = fmt.Sprintf("chatterino-api-cache/%s link-resolver (bot)", version.Version)
		// TODO: Use twitter-tweet-cache-duration?
		cacheDur = time.Hour * 24
	}

	resp, err := resolver.RequestGETWithHeaders(requestUrl.String(), extraHeaders)
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
		log.Infow("Skipping url because of status code", "url", resp.Request.URL, "status", resp.StatusCode)
		return staticresponse.SNoLinkInfoFound.Return()
	}

	contentType := resp.Header.Get("Content-Type")
	for _, ctResolver := range l.contentTypeResolvers {
		if ctResolver.Check(ctx, contentType) {
			ttResponse, err := ctResolver.Run(ctx, r, resp)
			if err != nil {
				log.Errorw("error running ContentTypeResolver",
					"resolver", ctResolver.Name(),
					"err", err,
				)

				return utils.MarshalNoDur(&resolver.Response{
					Status:  http.StatusInternalServerError,
					Message: "ContentTypeResolver error " + resolver.CleanResponse(err.Error()),
				})
			}

			return utils.MarshalNoDur(ttResponse)
		}
	}

	// Fallback to parsing via goquery
	limiter := &resolver.WriteLimiter{Limit: l.maxContentLength}
	doc, err := goquery.NewDocumentFromReader(io.TeeReader(resp.Body, limiter))
	if err != nil {
		body, bodyErr := io.ReadAll(resp.Body)
		log.Errorw("Bad body?", "body", body, "bodyErr", bodyErr, "err", err, "url", requestUrl)
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

	if thumbnail.IsSupportedThumbnailType(contentType) {
		response.Thumbnail = utils.FormatThumbnailURL(l.baseURL, r, resp.Request.URL.String())
	}

	finalData, finalErr := json.Marshal(response)
	return finalData, nil, nil, cacheDur, finalErr
}
