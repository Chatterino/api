package defaultresolver

import (
	"context"
	"html"
	"net/http"

	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

type ContentTypeResolver interface {
	Check(ctx context.Context, contentType string) bool
	Run(ctx context.Context, req *http.Request, resp *http.Response) (*resolver.Response, error)
	Name() string
}

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
