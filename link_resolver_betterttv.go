package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	errInvalidBTTVEmotePath = errors.New("invalid BetterTTV emote path")

	noBetterTTVEmoteWithThisHashFound = &LinkResolverResponse{
		Status:  http.StatusNotFound,
		Message: "No BetterTTV emote with this hash found",
	}
)

const betterttvEmoteTooltip = `<div style="text-align: left;">
<b>{{.Code}}</b><br>
<b>{{.Type}} BetterTTV Emote</b><br>
<b>By:</b> {{.Uploader}}</div>`

type betterttvEmoteTooltipData struct {
	Code     string
	Type     string
	Uploader string
}

const (
	betterttvThumbnailFormat = "https://cdn.betterttv.net/emote/%s/3x"
	betterttvEmoteAPIURL     = "https://api.betterttv.net/3/emotes/%s"
)

func init() {
	tooltipTemplate, err := template.New("betterttvEmoteTooltip").Parse(betterttvEmoteTooltip)
	if err != nil {
		log.Println("Error initialization BTTV Emotes tooltip template:", err)
		return
	}

	load := func(emoteHash string, r *http.Request) (interface{}, error, time.Duration) {
		type BTTVEmoteAPIResponse struct {
			ID             string    `json:"id"`
			Code           string    `json:"code"`
			ImageType      string    `json:"imageType"`
			CreatedAt      time.Time `json:"createdAt"`
			UpdatedAt      time.Time `json:"updatedAt"`
			Global         bool      `json:"global"`
			Live           bool      `json:"live"`
			Sharing        bool      `json:"sharing"`
			ApprovalStatus string    `json:"approvalStatus"`
			User           struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				DisplayName string `json:"displayName"`
				ProviderID  string `json:"providerId"`
			} `json:"user"`
		}

		apiURL := fmt.Sprintf(betterttvEmoteAPIURL, emoteHash)
		thumbnailURL := fmt.Sprintf(betterttvThumbnailFormat, emoteHash)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv http request creation error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

		resp, err := httpClient.Do(req)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv http request error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
			return noBetterTTVEmoteWithThisHashFound, nil, noSpecialDur
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv http body read error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		var jsonResponse BTTVEmoteAPIResponse
		if err := json.Unmarshal(body, &jsonResponse); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv api unmarshal error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		data := betterttvEmoteTooltipData{
			Code:     jsonResponse.Code,
			Type:     "Shared",
			Uploader: jsonResponse.User.DisplayName,
		}

		if jsonResponse.Global {
			data.Type = "Global"
		}

		var tooltip bytes.Buffer
		if err := tooltipTemplate.Execute(&tooltip, data); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "youtube template error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:    200,
			Tooltip:   tooltip.String(),
			Thumbnail: thumbnailURL,
		}, nil, noSpecialDur
	}

	cache := newLoadingCache("betterttv_emotes", load, 1*time.Hour)
	emotePathRegex := regexp.MustCompile(`/emotes/([a-f0-9]+)`)

	// Find clips that look like https://clips.twitch.tv/SlugHere
	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			if !strings.HasSuffix("."+url.Host, ".betterttv.com") {
				return false
			}

			if !emotePathRegex.MatchString(url.Path) {
				return false
			}

			return true
		},
		run: func(url *url.URL) ([]byte, error) {
			matches := emotePathRegex.FindStringSubmatch(url.Path)
			if len(matches) != 2 {
				return nil, errInvalidBTTVEmotePath
			}

			emoteHash := matches[1]

			apiResponse := cache.Get(emoteHash, nil)
			return json.Marshal(apiResponse)
		},
	})
}
