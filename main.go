package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	cache "github.com/patrickmn/go-cache"
)

var kvCache *cache.Cache
var firstRun = true

const offlineMode = false

func getData(url, key string) ([]byte, error) {
	if offlineMode {
		raw, err := ioutil.ReadFile("./cached/" + key)
		if err != nil {
			return nil, err
		}

		return raw, nil
	}

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
	fmt.Println("Fetched!")
	return body, nil
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
	Type        string `json:"type"`
	Custom      bool   `json:"custom"`
}

var emoteSets map[string]*EmoteSet
var emoteSetMutex sync.Mutex

func addEmoteSet(emoteSetID, channelName, channelID, setType string) {
	emoteSets[emoteSetID] = &EmoteSet{
		ChannelName: channelName,
		ChannelID:   channelID,
		Type:        setType,
		Custom:      true,
	}
}

func refreshEmoteSetCache() {
	if firstRun {
		emoteSetMutex.Lock()
		defer emoteSetMutex.Unlock()
	}

	data, err := getData("https://twitchemotes.com/api_cache/v3/sets.json", "twitchemotes:sets")
	if err != nil {
		panic(err)
	}

	if !firstRun {
		emoteSetMutex.Lock()
		defer emoteSetMutex.Unlock()
	}

	firstRun = false
	emoteSets = make(map[string]*EmoteSet)

	err = json.Unmarshal(data, &emoteSets)
	if err != nil {
		panic(err)
	}

	for k, _ := range emoteSets {
		emoteSets[k].Type = "sub"
	}

	addEmoteSet("13985", "evohistorical2015", "129284508", "sub")

	fmt.Println("Refreshed emote sets")

	time.AfterFunc(30*time.Minute, refreshEmoteSetCache)
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	setID := vars["setID"]
	emoteSetMutex.Lock()
	defer emoteSetMutex.Unlock()
	data, err := json.Marshal(emoteSets[setID])
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: returning data %s\n", setID, data)
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

var host = flag.String("h", ":1234", "host of server")
var prefix = flag.String("p", "", "optional prefix")

func main() {
	flag.Parse()
	go refreshEmoteSetCache()
	kvCache = cache.New(30*time.Minute, 10*time.Minute)

	router := mux.NewRouter()

	sr := router.PathPrefix(*prefix).Subrouter()

	sr.HandleFunc("/twitchemotes/sets", cacheRequest("https://twitchemotes.com/api_cache/v3/sets.json", "twitchemotes:sets", 30*time.Minute)).Methods("GET")
	sr.HandleFunc("/twitchemotes/subscriber", cacheRequest("https://twitchemotes.com/api_cache/v3/subscriber.json", "twitchemotes:subscriber", 30*time.Minute)).Methods("GET")

	sr.HandleFunc("/twitchemotes/set/{setID}/", setHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(*host, router))
}
