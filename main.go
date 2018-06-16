package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cache "github.com/patrickmn/go-cache"
)

var kvCache *cache.Cache

func getData(url, key string) ([]byte, error) {
	raw, err := ioutil.ReadFile("./cached/" + key)
	if err != nil {
		return nil, err
	}

	return raw, nil

	/*
		resp, err := http.Get(url)
		log.Printf("Fetching %s live...", url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		kvCache.Set(key, body, cacheDuration)
		return body, nil
	*/
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

type EmoteSet struct {
	ChannelName string `json:"channel_name"`
	ChannelID   string `json:"channel_id"`
}

var emoteSets map[string]EmoteSet

func refreshEmoteSetCache() {
	data, err := getData("https://twitchemotes.com/api_cache/v3/sets.json", "twitchemotes:sets")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &emoteSets)
	if err != nil {
		panic(err)
	}

	time.AfterFunc(30*time.Minute, refreshEmoteSetCache)
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	setID := vars["setID"]
	data, err := json.Marshal(emoteSets[setID])
	if err != nil {
		panic(err)
	}
	fmt.Printf("returning data %s", data)
	w.Write(data)
}

var host = flag.String("h", ":1234", "host of server")

func main() {
	flag.Parse()
	go refreshEmoteSetCache()
	emoteSets = make(map[string]EmoteSet)
	kvCache = cache.New(30*time.Minute, 10*time.Minute)

	router := mux.NewRouter()

	router.HandleFunc("/twitchemotes/sets", cacheRequest("https://twitchemotes.com/api_cache/v3/sets.json", "twitchemotes:sets", 30*time.Minute)).Methods("GET")
	router.HandleFunc("/twitchemotes/subscriber", cacheRequest("https://twitchemotes.com/api_cache/v3/subscriber.json", "twitchemotes:subscriber", 30*time.Minute)).Methods("GET")

	router.HandleFunc("/twitchemotes/set/{setID}/", setHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(*host, router))
}
