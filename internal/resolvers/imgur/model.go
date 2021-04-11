package imgur

type miniImage struct {
	Title       string
	Description string
	UploadDate  string

	Nsfw     bool
	Animated bool
	Album    bool

	mimeType string

	// size of image in bytes
	size int

	// TODO: Name?
	// TODO: Section?

	// Direct link to the image, used as a thumbnail
	Link string
}

const imageTooltip = `<div style="text-align: left;">` +
	`{{ if .Title }}<li><b>Title:</b> {{ .Title }}</li>{{ end }}` +
	`{{ if .Description }}<li><b>Description:</b> {{.Description}}</li>{{ end }}` +
	`<li><b>Uploaded:</b> {{.UploadDate}}</li>` +
	`{{ if .Nsfw }}<li><b><span style="color: red;">NSFW</span></b></li>{{ end }}` +
	`{{ if .Animated }}<li><b><span style="color: red;">ANIMATED</span></b></li>{{ end }}` +
	`</div>`
