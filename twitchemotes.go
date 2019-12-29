package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type TwitchEmotesError struct {
	Status int
	Error  error
}

type TwitchEmotesErrorResponse struct {
	Status  int
	Message string
}

type EmoteSet struct {
	ChannelName string `json:"channel_name"`
	ChannelID   string `json:"channel_id"`
	Type        string `json:"type"`
	Tier        int    `json:"tier"`
	Custom      bool   `json:"custom"`
}

var (
	errInvalidEmoteID = errors.New("invalid emote id")
)

var customEmoteSets map[string][]byte = make(map[string][]byte)

var twitchemotesCache *loadingCache = newLoadingCache("twitchemotes", doTwitchemotesRequest, 30*time.Minute)

func addEmoteSet(emoteSetID, channelName, channelID, setType string) {
	b, err := json.Marshal(&EmoteSet{
		ChannelName: channelName,
		ChannelID:   channelID,
		Type:        setType,
		Custom:      true,
	})
	if err != nil {
		panic(err)
	}
	customEmoteSets[emoteSetID] = b
}

func init() {
	addEmoteSet("13985", "evohistorical2015", "129284508", "sub")
}

func setsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement multiset-fetcher and in future version of Chatterino which sends a list of sets instead of one per request
	w.Write([]byte("{\"error\": \"not implemented\"}"))
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	setID := vars["setID"]

	// 1. Check our "custom" responses
	if v, ok := customEmoteSets[setID]; ok {
		_, err := w.Write(v)
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	// 2. Cache a request from twitchemotes.com
	response := twitchemotesCache.Get(setID)

	switch v := response.(type) {
	case []byte:
		_, err := w.Write(v)
		if err != nil {
			log.Println("Error writing response:", err)
		}

	case *TwitchEmotesError:
		w.WriteHeader(v.Status)
		data, err := json.Marshal(&TwitchEmotesErrorResponse{
			Status:  v.Status,
			Message: v.Error.Error(),
		})
		if err != nil {
			log.Println("Error marshalling twitch emotes error response:", err)
		}
		_, err = w.Write(data)
		if err != nil {
			log.Println("Error writing response:", err)
		}
	}
}

func doTwitchemotesRequest(setID string) (interface{}, error, time.Duration) {
	url := fmt.Sprintf("https://api.twitchemotes.com/api/v4/sets?id=%s", setID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &TwitchEmotesError{
			Error:  err,
			Status: 500,
		}, nil, 0
	}

	req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

	resp, err := httpClient.Do(req)
	if err != nil {
		return &TwitchEmotesError{
			Error:  err,
			Status: 500,
		}, nil, 0
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &TwitchEmotesError{
			Error:  err,
			Status: 500,
		}, nil, 0
	}
	var emoteSets []EmoteSet
	err = json.Unmarshal(body, &emoteSets)
	if err != nil {
		return &TwitchEmotesError{
			Error:  err,
			Status: 500,
		}, nil, 0
	}

	if len(emoteSets) == 0 {
		return &TwitchEmotesError{
			Error:  errInvalidEmoteID,
			Status: 404,
		}, nil, 0
	}

	if len(emoteSets) > 1 {
		log.Println("Unhandled long emote set for emote set", setID)
	}

	r := &emoteSets[0]
	r.Type = "sub"

	return marshalNoDur(&emoteSets[0])
}

func handleTwitchEmotes(router *mux.Router) {
	router.HandleFunc("/twitchemotes/set/{setID}/", setHandler).Methods("GET")

	router.HandleFunc("/twitchemotes/sets/", setsHandler).Methods("GET")
}
