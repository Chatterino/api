package livestreamfails

import "time"

type TooltipData struct {
	NSFW         bool
	Title        string
	Category     string
	RedditScore  string
	Platform     string
	StreamerName string
	CreationDate string
}

type Resize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Output struct {
	Format string `json:"format"`
}

type LivestreamFailsThumbnailRequest struct {
	Input  string `json:"input"`
	Resize Resize `json:"resize"`
	Output Output `json:"output"`
}

type LivestreamfailsAPIResponse struct {
	Category struct {
		Label string `json:"label"`
	} `json:"category"`
	CreatedAt      time.Time `json:"createdAt"`
	ImageID        string    `json:"imageId"`
	IsNSFW         bool      `json:"isNSFW"`
	Label          string    `json:"label"`
	RedditScore    int       `json:"redditScore"`
	SourcePlatform string    `json:"sourcePlatform"`
	Streamer       struct {
		Label string `json:"label"`
	} `json:"streamer"`
}
