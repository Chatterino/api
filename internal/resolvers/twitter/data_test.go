package twitter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

var (
	users  = map[string]*TwitterUserApiResponse{}
	tweets = map[string]*TweetApiResponse{}
)

func init() {
	users["pajlada"] = &TwitterUserApiResponse{
		Name:            "PAJLADA",
		Username:        "pajlada",
		Description:     "Cool memer",
		Followers:       69,
		ProfileImageUrl: "https://pbs.twimg.com/profile_images/1385924241619628033/fW7givJA_400x400.jpg",
	}

	// Tweet with no entities
	tweets["1507648130682077194"] = &TweetApiResponse{
		Text: "Digging a hole",
		User: APIUser{
			Name:     "PAJLADA",
			Username: "pajlada",
		},
		Likes:     69,
		Retweets:  420,
		Timestamp: "Sat Mar 26 17:15:50 +0200 2022",
	}

	// Tweet with entities
	tweets["1506968434134953986"] = &TweetApiResponse{
		Text: "",
		User: APIUser{
			Name:     "PAJLADA",
			Username: "pajlada",
		},
		Likes:     69,
		Retweets:  420,
		Timestamp: "Sat Mar 26 17:15:50 +0200 2022",
		Entities: APIEntities{
			Media: []APIEntitiesMedia{
				{
					Url: "https://pbs.twimg.com/media/FOnTzeQWUAMU6L1?format=jpg&name=medium",
				},
			},
		},
	}

	// Tweet with poorly formatted timestamp
	tweets["1505121705290874881"] = &TweetApiResponse{
		Text: "Bad timestamp",
		User: APIUser{
			Name:     "PAJLADA",
			Username: "pajlada",
		},
		Likes:     420,
		Retweets:  69,
		Timestamp: "asdasd",
	}
}

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/1.1/users/show.json", func(w http.ResponseWriter, r *http.Request) {
		screenName := r.URL.Query().Get("screen_name")

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
	r.Get("/1.1/statuses/show.json", func(w http.ResponseWriter, r *http.Request) {
		tweetID := r.URL.Query().Get("id")

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
