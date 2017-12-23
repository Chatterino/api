package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cache "github.com/patrickmn/go-cache"
)

var kvCache *cache.Cache

func cacheRequest(url, key string, cacheDuration time.Duration) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, found := kvCache.Get(key)
		if found {
			w.Write(data.([]byte))
		} else {
			resp, err := http.Get(url)
			log.Printf("Fetching %s live...", url)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}
			kvCache.Set(key, body, cacheDuration)
			w.Write(body)
		}
	}
}

func main() {
	kvCache = cache.New(30*time.Minute, 10*time.Minute)

	router := mux.NewRouter()
	router.HandleFunc("/twitchemotes/sets", cacheRequest("https://twitchemotes.com/api_cache/v3/sets.json", "twitchemotes:sets", 30*time.Minute)).Methods("GET")

	log.Fatal(http.ListenAndServe(":1234", router))
}
