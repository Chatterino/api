package defaultresolver

import (
	"bytes"
	"context"
	"html"
	"html/template"
	"io"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/validate"
)

const templateString = `<div style="text-align: left;">
<b>PDF File</b><br>
{{if .Title}}<b>Title:</b> {{.Title}}<br>{{end}}
{{if .Author}}<b>Author:</b> {{.Author}}<br>{{end}}
<span style="color: #808892;">
{{.PageCount}} pages{{if .CreationDate}}&nbsp;•&nbsp;{{.CreationDate}}{{end}}
</span>
</div>
`

var pdfTooltipTemplate = template.Must(template.New("pdfTooltipTemplate").Parse(templateString))

type pdfTooltipData struct {
	Title        string
	Author       string
	PageCount    int
	CreationDate string
}

type PDFResolver struct {
	baseURL          string
	maxContentLength uint64
}

func (r *PDFResolver) Check(ctx context.Context, contentType string) bool {
	return contentType == "application/pdf"
}

func (r *PDFResolver) Run(ctx context.Context, req *http.Request, resp *http.Response) (*resolver.Response, error) {
	log := logger.FromContext(ctx)

	limiter := resolver.WriteLimiter{Limit: r.maxContentLength}
	limitedReader := io.TeeReader(resp.Body, &limiter)
	buffer, err := io.ReadAll(limitedReader)
	if err != nil {
		log.Errorw("error reading response body", "err", err)
		return nil, err
	}

	readSeeker := bytes.NewReader(buffer)

	pdfCtx, err := api.ReadContext(readSeeker, pdfcpu.NewDefaultConfiguration())
	if err != nil {
		log.Errorw("error reading pdf context", "err", err)
		return nil, err
	}

	if err = validate.XRefTable(pdfCtx.XRefTable); err != nil {
		log.Errorw("error validating XRefTable", "err", err)
		return nil, err
	}

	dtString := ""
	if creationDt, ok := pdfcpu.DateTime(pdfCtx.CreationDate, true); ok {
		dtString = humanize.CreationDate(creationDt)
	}

	ttData := pdfTooltipData{
		Title:        html.EscapeString(humanize.Title(pdfCtx.Title)),
		Author:       html.EscapeString(humanize.Title(pdfCtx.Author)),
		PageCount:    pdfCtx.PageCount,
		CreationDate: dtString,
	}

	var tooltip bytes.Buffer
	if err := pdfTooltipTemplate.Execute(&tooltip, ttData); err != nil {
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

func (r *PDFResolver) Name() string {
	return "PDFResolver"
}

func NewPDFResolver(baseURL string, maxContentLength uint64) *PDFResolver {
	return &PDFResolver{
		baseURL:          baseURL,
		maxContentLength: maxContentLength,
	}
}
