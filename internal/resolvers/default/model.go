package defaultresolver

import (
	"html"
	"net/http"

	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/PuerkitoBio/goquery"
)

type tooltipData struct {
	URL         string
	Title       string
	Description string
	ImageSrc    string
}

func (d *tooltipData) Truncate() {
	d.Title = humanize.Title(d.Title)
	d.Description = humanize.Description(d.Description)
}

func (d *tooltipData) Sanitize() {
	d.Title = html.EscapeString(d.Title)
	d.Description = html.EscapeString(d.Description)
}

// does this really fit in model?
func (dr *R) defaultTooltipData(doc *goquery.Document, r *http.Request, resp *http.Response) tooltipData {
	data := tooltipMetaFields(dr.cfg.BaseURL, doc, r, resp, tooltipData{
		URL: resolver.CleanResponse(resp.Request.URL.String()),
	})

	if data.Title == "" {
		data.Title = doc.Find("title").First().Text()
	}

	return data
}

const (
	defaultTooltipString = `<div style="text-align: left;">
{{if .Title}}
<b>{{.Title}}</b><hr>
{{end}}
{{if .Description}}
<span>{{.Description}}</span><hr>
{{end}}
<b>URL:</b> {{.URL}}</div>`
)
