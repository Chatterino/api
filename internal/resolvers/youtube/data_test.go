package youtube

import youtubeAPI "google.golang.org/api/youtube/v3"

var (
	videos = map[string]*youtubeAPI.VideoListResponse{}
)

func init() {
	videos["foobar"] = &youtubeAPI.VideoListResponse{
		Items: []*youtubeAPI.Video{
			{
				ContentDetails: &youtubeAPI.VideoContentDetails{
					ContentRating: &youtubeAPI.ContentRating{
						YtRating: "ytAgeRestricted",
					},
					Duration: "PT#5#2",
				},
				Snippet: &youtubeAPI.VideoSnippet{
					Title:        "Video Title",
					ChannelTitle: "Channel Title",
					PublishedAt:  "2019-10-12T07:20:50.52Z",
					Thumbnails: &youtubeAPI.ThumbnailDetails{
						Default: &youtubeAPI.Thumbnail{
							Url: "https://example.com/thumbnail.png",
						},
					},
				},
				Statistics: &youtubeAPI.VideoStatistics{
					ViewCount:    50,
					LikeCount:    10,
					CommentCount: 5,
				},
			},
		},
	}

	videos["mediumtn"] = &youtubeAPI.VideoListResponse{
		Items: []*youtubeAPI.Video{
			{
				ContentDetails: &youtubeAPI.VideoContentDetails{
					ContentRating: &youtubeAPI.ContentRating{
						YtRating: "ytAgeRestricted",
					},
					Duration: "PT#5#2",
				},
				Snippet: &youtubeAPI.VideoSnippet{
					Title:        "Video Title",
					ChannelTitle: "Channel Title",
					PublishedAt:  "2019-10-12T07:20:50.52Z",
					Thumbnails: &youtubeAPI.ThumbnailDetails{
						Default: &youtubeAPI.Thumbnail{
							Url: "https://example.com/thumbnail.png",
						},
						Medium: &youtubeAPI.Thumbnail{
							Url: "https://example.com/medium.png",
						},
					},
				},
				Statistics: &youtubeAPI.VideoStatistics{
					ViewCount:    50,
					LikeCount:    10,
					CommentCount: 5,
				},
			},
		},
	}

	videos["404"] = &youtubeAPI.VideoListResponse{
		Items: []*youtubeAPI.Video{},
	}

	// Unrealistic response
	videos["toomany"] = &youtubeAPI.VideoListResponse{
		Items: []*youtubeAPI.Video{
			{},
			{},
		},
	}

	videos["unavailable"] = &youtubeAPI.VideoListResponse{
		Items: []*youtubeAPI.Video{
			{
				ContentDetails: nil,
			},
		},
	}
}
