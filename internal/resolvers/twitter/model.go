package twitter

type TweetApiResponse struct {
	ID        string `json:"id_str"`
	Text      string `json:"full_text"`
	Timestamp string `json:"created_at"`
	Likes     uint64 `json:"favorite_count"`
	Retweets  uint64 `json:"retweet_count"`
	User      struct {
		Name            string `json:"name"`
		Username        string `json:"screen_name"`
		ProfileImageUrl string `json:"profile_image_url_https"`
	} `json:"user"`
	Entities struct {
		Media []struct {
			Url string `json:"media_url_https"`
		} `json:"media"`
	} `json:"entities"`
}

type tweetTooltipData struct {
	Text      string
	Name      string
	Username  string
	Timestamp string
	Likes     string
	Retweets  string
	Thumbnail string
}

type TwitterUserApiResponse struct {
	Name            string `json:"name"`
	Username        string `json:"screen_name"`
	Description     string `json:"description"`
	Followers       uint64 `json:"followers_count"`
	ProfileImageUrl string `json:"profile_image_url_https"`
}

type twitterUserTooltipData struct {
	Name        string
	Username    string
	Description string
	Followers   string
	Thumbnail   string
}

const (
	templateStringTweet = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<span style="white-space: pre-wrap; word-wrap: break-word;">
{{.Text}}
</span>
<span style="color: #808892;">{{.Likes}} likes&nbsp;•&nbsp;{{.Retweets}} retweets&nbsp;•&nbsp;{{.Timestamp}}</span>
</div>
`

	templateStringTwitterUser = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<span style="white-space: pre-wrap; word-wrap: break-word;">
{{.Description}}
</span>
<span style="color: #808892;">{{.Followers}} followers</span>
</div>
`
)
