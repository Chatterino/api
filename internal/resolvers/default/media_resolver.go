package defaultresolver

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const mediaTooltipTemplateString = `<div style="text-align: left;">
<b>Media File</b>
{{if .MediaType}}<br><b>Type:</b> {{.MediaType}}{{end}}
{{if .Size}}<br><b>Size:</b> {{.Size}}{{end}}
</div>
`

var mediaTooltipTemplate = template.Must(template.New("mediaTooltipTemplate").Parse(mediaTooltipTemplateString))

type mediaTooltipData struct {
	MediaType string
	Size      string
}

type MediaResolver struct {
	baseURL string
}

func (r *MediaResolver) Check(ctx context.Context, contentType string) bool {
	spl := strings.Split(contentType, "/")
	switch spl[0] {
	case "video", "audio":
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
		MediaType: fmt.Sprintf("%s (%s)",
			cases.Title(language.English).String(spl[0]),
			strings.ToUpper(extensionFromMime(mimeType)),
		),
		Size: size,
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
		case "mpeg":
			return "mp3"
		}
	case "video":
		switch s2 {
		case "avi", "x-msvideo":
			return "avi"
		case "mp4":
			return "mp4"
		case "quicktime":
			return "mov"
		}
	}

	// this returns weird extensions for some mime types
	// so it's only used as a backup.
	// video/mp4 returns f4v for example.
	types, _ := mime.ExtensionsByType(mimeType)
	if len(types) > 0 {
		ext := types[0]
		if len(ext) > 1 {
			ext = ext[1:]
		}
		return ext
	}
	return "unknown"
}
