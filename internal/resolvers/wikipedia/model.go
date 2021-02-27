package wikipedia

// The `Thumbnail` and `Description` fields are declared as pointers because
// they are not strictly required by the schema and may be omitted for some
// pages. In these cases, the fields will be nil.
type wikipediaAPIResponse struct {
	Titles struct {
		Display string `json:"display"`
	} `json:"titles"`
	Extract   string `json:"extract"`
	Thumbnail *struct {
		URL string `json:"source"`
	} `json:"thumbnail"`
	Description *string `json:"description"`
}

type wikipediaTooltipData struct {
	Title        string
	Description  string
	Extract      string
	ThumbnailURL string
}

const wikipediaTooltip = `<div style="text-align: left;">
<b>{{.Title}}{{ if .Description }}&nbsp;•&nbsp;{{.Description}}{{ end }}</b><br>
{{.Extract}}
</div>
`
