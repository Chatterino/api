package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/go-chi/chi/v5"
	"github.com/nicklaw5/helix"
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

	customEmoteSets map[string][]byte = make(map[string][]byte)

	twitchemotesCache = cache.New("twitchemotes", doTwitchemotesRequest, time.Duration(30)*time.Minute)

	helixAPI *helix.Client
)

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
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte("{\"error\": \"not implemented\"}"))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	setID := chi.URLParam(r, "setID")
	w.Header().Set("Content-Type", "application/json")

	// 1. Check our "custom" responses
	if v, ok := customEmoteSets[setID]; ok {
		_, err := w.Write(v)
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	// 2. Cache a request from twitchemotes.com
	response := twitchemotesCache.Get(setID, nil)

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

func doTwitchemotesRequest(setID string, r *http.Request) (interface{}, time.Duration, error) {
	if helixAPI == nil {
		// TODO: Return specific error saying this server isn't able to handle requests
		return &TwitchEmotesError{
			Error:  errInvalidEmoteID,
			Status: 404,
		}, 0, nil
	}

	params := &helix.GetEmoteSetsParams{
		EmoteSetIDs: []string{
			setID,
		},
	}
	resp, err := helixAPI.GetEmoteSets(params)
	if err != nil {
		return &TwitchEmotesError{
			Error:  err,
			Status: 500,
		}, 0, nil
	}

	fmt.Println(resp.Data.Emotes)
	fmt.Printf("%#v\n", resp.Data)

	// emoteSet.ChannelID = resp.Data.Emotes.

	// err = json.Unmarshal(body, &emoteSets)
	// if err != nil {
	// 	return &TwitchEmotesError{
	// 		Error:  err,
	// 		Status: 500,
	// 	}, 0, nil
	// }

	// if len(emoteSets) == 0 {
	// 	return &TwitchEmotesError{
	// 		Error:  errInvalidEmoteID,
	// 		Status: 404,
	// 	}, 0, nil
	// }

	// if len(emoteSets) > 1 {
	// 	log.Println("Unhandled long emote set for emote set", setID)
	// }

	// return utils.MarshalNoDur(&emoteSets[0])
	return &TwitchEmotesError{
		Error:  errInvalidEmoteID,
		Status: 404,
	}, 0, nil
}

func handleTwitchEmotes(router *chi.Mux, helixClient *helix.Client) {
	helixAPI = helixClient
	router.Get("/twitchemotes/set/{setID}/", setHandler)

	router.Get("/twitchemotes/sets/", setsHandler)
}
