package youtube

type youtubeVideoTooltipData struct {
	Title        string
	ChannelTitle string
	Duration     string
	PublishDate  string
	Views        string
	LikeCount    string
}

type youtubeChannelTooltipData struct {
	Title       string
	JoinedDate  string
	Subscribers string
	Views       string
}

const (
	templateStringYoutubeVideo = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Channel:</b> {{.ChannelTitle}}
<br><b>Duration:</b> {{.Duration}}
<br><b>Published:</b> {{.PublishDate}}
<br><b>Views:</b> {{.Views}}
<br><span style="color: #2ecc71;">{{.LikeCount}} likes</span>
</div>
`

	templateStringYoutubeChannel = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Joined Date:</b> {{.JoinedDate}}
<br><b>Subscribers:</b> {{.Subscribers}}
<br><b>Views:</b> {{.Views}}
</div>
`
)
