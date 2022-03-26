package twitter

import (
	"time"

	"github.com/Chatterino/api/pkg/humanize"
)

func buildTweetTooltip(tweet *TweetApiResponse) *tweetTooltipData {
	data := &tweetTooltipData{}
	data.Text = tweet.Text
	data.Name = tweet.User.Name
	data.Username = tweet.User.Username
	data.Likes = humanize.Number(tweet.Likes)
	data.Retweets = humanize.Number(tweet.Retweets)

	// TODO: what time format is this exactly? can we move to humanize a la CreationDteRFC3339?
	timestamp, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", tweet.Timestamp)
	if err != nil {
		data.Timestamp = ""
	} else {
		data.Timestamp = humanize.CreationDateTime(timestamp)
	}

	if len(tweet.Entities.Media) > 0 {
		// If tweet contains an image, it will be used as thumbnail
		data.Thumbnail = tweet.Entities.Media[0].Url
	}

	return data
}

func buildTwitterUserTooltip(user *TwitterUserApiResponse) *twitterUserTooltipData {
	data := &twitterUserTooltipData{}
	data.Name = user.Name
	data.Username = user.Username
	data.Description = user.Description
	data.Followers = humanize.Number(user.Followers)
	data.Thumbnail = user.ProfileImageUrl

	return data
}
