package wikipedia

type wikipediaAPITitles struct {
	Normalized string `json:"normalized"`
}

type wikipediaAPIThumbnail struct {
	URL string `json:"source"`
}

// The `Thumbnail` and `Description` fields are declared as pointers because
// they are not strictly required by the schema and may be omitted for some
// pages. In these cases, the fields will be nil.
type wikipediaAPIResponse struct {
	Titles      wikipediaAPITitles     `json:"titles"`
	Extract     string                 `json:"extract"`
	Thumbnail   *wikipediaAPIThumbnail `json:"thumbnail"`
	Description *string                `json:"description"`
}

type wikipediaTooltipData struct {
	Title        string
	Description  string
	Extract      string
	ThumbnailURL string
}

const wikipediaTooltip = `<div style="text-align: left;">` +
	`<b>{{.Title}}{{ if .Description }}&nbsp;•&nbsp;{{.Description}}{{ end }}</b><br>` +
	`{{.Extract}}` +
	`</div>`
