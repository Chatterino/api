package supinic

import "time"

type TooltipData struct {
	ID         int
	Name       string
	AuthorName string
	Tags       string
	Duration   string
}

type TrackData struct {
	ID          int       `json:"id"`
	Link        string    `json:"code"` // Youtube ID/link
	Name        string    `json:"name"`
	VideoType   int       `json:"videoType"`
	TrackType   string    `json:"trackType"`
	Duration    float32   `json:"duration"`
	Available   bool      `json:"available"`
	PublishedAt time.Time `json:"published"`
	Notes       string    `json:"notes"`
	AddedBy     string    `json:"addedBy"`
	ParsedLink  string    `json:"parsedLink"`
	Tags        []string  `json:"tags"`
	Authors     []struct {
		ID   int    `json:"ID"`
		Name string `json:"name"`
		Role string `json:"role"`
	} `json:"authors"`
}

type TrackListAPIResponse struct {
	Data TrackData `json:"data"`
}
