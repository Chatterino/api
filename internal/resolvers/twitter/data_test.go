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
