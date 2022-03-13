package defaultresolver

import (
	"html"

	"github.com/Chatterino/api/pkg/humanize"
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
