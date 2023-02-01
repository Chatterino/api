package defaultresolver

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

const mediaTooltipTemplateString = `<div style="text-align: left;">
<b>Media File</b>
{{if .MediaType}}<br><b>Type:</b> {{.MediaType}}{{end}}
{{if .Extension}}<br><b>Extension:</b> {{.Extension}}{{end}}
{{if .Size}}<br><b>Size:</b> {{.Size}}{{end}}
</div>
`

var mediaTooltipTemplate = template.Must(template.New("mediaTooltipTemplate").Parse(mediaTooltipTemplateString))

type mediaTooltipData struct {
	MediaType string
	Extension string
	Size      string
}

type MediaResolver struct {
	baseURL string
}

func (r *MediaResolver) Check(ctx context.Context, contentType string) bool {
	spl := strings.Split(contentType, "/")
	switch spl[0] {
	case "video", "audio", "application":
		return true
	}
	return false
}

func (r *MediaResolver) Run(ctx context.Context, req *http.Request, resp *http.Response) (*resolver.Response, error) {
	mimeType := resp.Header.Get("Content-Type")
	spl := strings.Split(mimeType, "/")

	size := ""
	reportedSize := resp.ContentLength
	if reportedSize > 0 {
		size = humanize.Bytes(uint64(reportedSize))
	}

	ttData := mediaTooltipData{
		MediaType: strings.Title(spl[0]),
		Extension: extensionFromMime(mimeType),
		Size:      size,
	}

	var tooltip bytes.Buffer
	if err := mediaTooltipTemplate.Execute(&tooltip, ttData); err != nil {
		return nil, err
	}

	targetURL := resp.Request.URL.String()
	response := &resolver.Response{
		Status:    http.StatusOK,
		Link:      targetURL,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: utils.FormatThumbnailURL(r.baseURL, req, targetURL),
	}

	return response, nil
}

func (r *MediaResolver) Name() string {
	return "MediaResolver"
}

func NewMediaResolver(baseURL string) *MediaResolver {
	return &MediaResolver{
		baseURL: baseURL,
	}
}

func extensionFromMime(mimeType string) string {
	spl := strings.Split(mimeType, "/")
	if len(spl) < 2 {
		return ""
	}
	s1, s2 := spl[0], spl[1]
	switch s1 {
	case "audio":
		switch s2 {
		case "wav", "x-wav":
			return "wav"
		default:
			return "mp3"
		}
	case "video":
		switch s2 {
		case "avi":
			return "avi"
		case "quicktime":
			return "mov"
		default:
			return "mp4"
		}
	case "application":
		switch s2 {
		case "json":
			return "json"
		case "x-gzip":
			return "gz"
		case "javascript", "x-javascript", "ecmascript":
			return "js"
		case "pdf":
			return "pdf"
		case "xml":
			return "xml"
		case "x-compressed", "x-zip-compressed", "zip":
			return "zip"
		}
	}
	return "Unknown"
}
