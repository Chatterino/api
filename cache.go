package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var kvCache *cache.Cache

func init() {
	kvCache = cache.New(30*time.Minute, 10*time.Minute)
}

func cacheRequest(url, key string, cacheDuration time.Duration) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, found := kvCache.Get(key)
		if found {
			log.Printf("Responding with cached %s", url)
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
