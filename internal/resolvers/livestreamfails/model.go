package livestreamfails

import "time"

type TooltipData struct {
	NSFW         bool
	Title        string
	RedditScore  int
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
		CreatedAt      time.Time  `json:"createdAt"`
		DeletedAt      *time.Time `json:"deletedAt"` // Can be "null"/nil
		ID             int        `json:"id"`
		ImageID        string     `json:"imageId"`
		IsNSFW         bool       `json:"isNSFW"`
		Label          string     `json:"label"`
		SourceID       string     `json:"sourceId"`
		SourceLink     string     `json:"sourceLink"`
		SourcePlatform string     `json:"sourcePlatform"`
		UpdatedAt      time.Time  `json:"updatedAt"`
	} `json:"category"`
	CategoryID            int        `json:"categoryId"`
	CreatedAt             time.Time  `json:"createdAt"`
	DeletedAt             *time.Time `json:"deletedAt"` // Can be "null"/nil
	ID                    int        `json:"id"`
	ImageID               string     `json:"imageId"`
	IsLegacy              bool       `json:"isLegacy"`
	IsNSFW                bool       `json:"isNSFW"`
	Label                 string     `json:"label"`
	RedditID              string     `json:"redditId"`
	RedditMirrorCommentID string     `json:"redditMirrorCommentId"`
	RedditScore           int        `json:"redditScore"`
	SourceID              string     `json:"sourceId"`
	SourceLink            string     `json:"sourceLink"`
	SourcePlatform        string     `json:"sourcePlatform"`
	Streamer              struct {
		CreatedAt      time.Time  `json:"createdAt"`
		DeletedAt      *time.Time `json:"deletedAt"` // Can be "null"/nil
		ID             int        `json:"id"`
		ImageID        string     `json:"imageId"`
		IsNSFW         bool       `json:"isNSFW"`
		Label          string     `json:"label"`
		SourceID       string     `json:"sourceId"`
		SourceLink     string     `json:"sourceLink"`
		SourcePlatform string     `json:"sourcePlatform"`
	} `json:"streamer"`
	StreamerID int       `json:"streamerId"`
	UpdatedAt  time.Time `json:"updatedAt"`
	VideoId    string    `json:"videoId"`
}
