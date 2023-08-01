package twitter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	users  = map[string]*TwitterUserApiResponse{}
	tweets = map[string]*TweetApiResponse{}
	embeds = map[string]*EmbedApiResponse{}
)

func init() {
	users["pajlada"] = &TwitterUserApiResponse{
		Data: []TwitterUserData{
			{
				Name:            "PAJLADA",
				Username:        "pajlada",
				Description:     "Cool memer",
				ProfileImageUrl: "https://pbs.twimg.com/profile_images/1385924241619628033/fW7givJA_400x400.jpg",
				PublicMetrics: TwitterUserPublicMetrics{
					Followers: 69,
				},
			},
		},
	}

	// Tweet with image media
	tweets["1506968434134953986"] = &TweetApiResponse{
		Data: Data{
			ID:        "1506968434134953986",
			CreatedAt: time.Date(2022, time.March, 26, 17, 15, 50, 0, time.UTC),
			PublicMetrics: PublicMetrics{
				RetweetCount: 420,
				LikeCount:    69,
			},
		},
		Includes: Includes{
			Users: []Users{
				{
					Name:     "PAJLADA",
					Username: "pajlada",
				},
			},
			Media: []Media{
				{
					URL: "https://pbs.twimg.com/media/FOnTzeQWUAMU6L1?format=jpg&name=medium",
				},
			},
		},
	}

	// Tweet with video media
	tweets["1507648130682077194"] = &TweetApiResponse{
		Data: Data{
			ID:        "1507648130682077194",
			Text:      "Digging a hole",
			CreatedAt: time.Date(2022, time.March, 26, 17, 15, 50, 0, time.UTC),
			PublicMetrics: PublicMetrics{
				RetweetCount: 420,
				LikeCount:    69,
			},
		},
		Includes: Includes{
			Users: []Users{
				{
					Name:     "PAJLADA",
					Username: "pajlada",
				},
			},
			Media: []Media{
				{
					Type:            "video",
					PreviewImageUrl: "https://pbs.twimg.com/ext_tw_video_thumb/1507648047609745413/pu/img/YZQAxKt-O68sKoXQ.jpg",
				},
			},
		},
	}

	// Tweet with no ID
	tweets["1505121705290874881"] = &TweetApiResponse{
		Data: Data{
			ID:        "",
			Text:      "No ID",
			CreatedAt: time.Date(2022, time.March, 26, 17, 15, 50, 0, time.UTC),
			PublicMetrics: PublicMetrics{
				RetweetCount: 69,
				LikeCount:    420,
			},
		},
		Includes: Includes{
			Users: []Users{
				{
					Name:     "PAJLADA",
					Username: "pajlada",
				},
			},
		},
	}

	// Embed with one photo
	embeds["1541102101782216706"] = &EmbedApiResponse{
		Text:      "Since I switched to rifle and 4:3 its been pretty easy. People calling me bad when they were already playing on the easiest difficulty🤣🤣🤣😎 https://t.co/jfbLdsPTwZ",
		ID:        "1541102101782216706",
		CreatedAt: time.Date(2022, time.June, 26, 16, 52, 28, 0, time.UTC),
		User: EmbedUser{
			Name:       "Sebastian Fors",
			ScreenName: "Forsen",
		},
		FavoriteCount:     4966,
		ConversationCount: 242,
		MediaDetails: []EmbedMediaDetail{
			{
				MediaUrl: "https://pbs.twimg.com/media/FWMYHYSWYAAyWMp.jpg",
			},
		},
	}

	// Embed without media
	embeds["1662863815787126787"] = &EmbedApiResponse{
		Text:      "Playing some more warlander today on stream! Free to play and join! Get it here https://t.co/rh3IBOvymE ! @PlayWarlander #Warlander",
		ID:        "1662863815787126787",
		CreatedAt: time.Date(2023, time.May, 28, 16, 50, 2, 0, time.UTC),
		User: EmbedUser{
			Name:       "Sebastian Fors",
			ScreenName: "Forsen",
		},
		FavoriteCount:     437,
		ConversationCount: 87,
		MediaDetails:      []EmbedMediaDetail{},
	}

	// Embed without id
	embeds["1662863815787126788"] = &EmbedApiResponse{
		Text:      "",
		ID:        "",
		CreatedAt: time.Date(2023, time.May, 28, 16, 50, 2, 0, time.UTC),
		User: EmbedUser{
			Name:       "Sebastian Fors",
			ScreenName: "Forsen",
		},
		FavoriteCount:     437,
		ConversationCount: 87,
		MediaDetails:      []EmbedMediaDetail{},
	}
}

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/2/users/by", func(w http.ResponseWriter, r *http.Request) {
		screenName := r.URL.Query().Get("usernames")

		var response *TwitterUserApiResponse
		var ok bool

		w.Header().Set("Content-Type", "application/json")

		if screenName == "bad" {
			w.Write([]byte("xD"))
		} else if screenName == "500" {
			http.Error(w, http.StatusText(500), 500)
			return
		} else if response, ok = users[screenName]; !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Write(b)
	})
	r.Get("/2/tweets/{id}", func(w http.ResponseWriter, r *http.Request) {
		tweetID := chi.URLParam(r, "id")

		var response *TweetApiResponse
		var ok bool

		w.Header().Set("Content-Type", "application/json")

		if tweetID == "bad" {
			w.Write([]byte("xD"))
		} else if tweetID == "500" {
			http.Error(w, http.StatusText(500), 500)
			return
		} else if response, ok = tweets[tweetID]; !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Write(b)
	})
	return httptest.NewServer(r)
}

func testEmbedServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/tweet-result", func(w http.ResponseWriter, r *http.Request) {
		tweetID := r.URL.Query().Get("id")

		var response *EmbedApiResponse
		var ok bool

		w.Header().Set("Content-Type", "application/json")

		if tweetID == "bad" {
			w.Write([]byte("xD"))
		} else if tweetID == "500" {
			http.Error(w, http.StatusText(500), 500)
			return
		} else if response, ok = embeds[tweetID]; !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, _ := json.Marshal(&response)

		w.Write(b)
	})
	return httptest.NewServer(r)
}
