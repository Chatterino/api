package youtube

type youtubeVideoTooltipData struct {
	Title         string
	ChannelTitle  string
	Duration      string
	PublishDate   string
	Views         string
	LikeCount     string
	AgeRestricted bool
}

type youtubeChannelTooltipData struct {
	Title       string
	JoinedDate  string
	Subscribers string
	Views       string
}
