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
)

func init() {
	const (
		emoteAPIURL = "https://api.betterttv.net/3/emotes/%s"

		thumbnailFormat = "https://cdn.betterttv.net/emote/%s/3x"

		tooltipTemplate = `<div style="text-align: left;">
<b>{{.Code}}</b><br>
<b>{{.Type}} BetterTTV Emote</b><br>
<b>By:</b> {{.Uploader}}</div>`
	)

	var (
		emoteNotFoundResponse = &LinkResolverResponse{
			Status:  http.StatusNotFound,
			Message: "No BetterTTV emote with this hash found",
		}
	)

	type TooltipData struct {
		Code     string
		Type     string
		Uploader string
	}

	type BetterTTVEmoteAPIResponse struct {
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

	tmpl, err := template.New("betterttvEmoteTooltip").Parse(tooltipTemplate)
	if err != nil {
		log.Println("Error initialization BTTV Emotes tooltip template:", err)
		return
	}

	load := func(emoteHash string, r *http.Request) (interface{}, error, time.Duration) {
		apiURL := fmt.Sprintf(emoteAPIURL, emoteHash)
		thumbnailURL := fmt.Sprintf(thumbnailFormat, emoteHash)

		// Create BetterTTV API request
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv http request creation error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

		// Execute BetterTTV API request
		resp, err := httpClient.Do(req)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv http request error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		defer resp.Body.Close()

		// Error out if the emote isn't found or something else went wrong with the request
		if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
			return emoteNotFoundResponse, nil, noSpecialDur
		}

		// Read response into a string
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv http body read error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		// Parse response into a predefined JSON blob (see BetterTTVEmoteAPIResponse struct above)
		var jsonResponse BetterTTVEmoteAPIResponse
		if err := json.Unmarshal(body, &jsonResponse); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "betterttv api unmarshal error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		// Build tooltip data from the API response
		data := TooltipData{
			Code:     jsonResponse.Code,
			Type:     "Shared",
			Uploader: jsonResponse.User.DisplayName,
		}

		if jsonResponse.Global {
			data.Type = "Global"
		}

		// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
		var tooltip bytes.Buffer
		if err := tmpl.Execute(&tooltip, data); err != nil {
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

	// BetterTTV hosts we're doing our smart things on
	betterttvDomains := map[string]struct{}{
		"betterttv.com":     {},
		"www.betterttv.com": {},
	}

	// Find links matching the BetterTTV direct emote link (e.g. https://betterttv.com/emotes/566ca06065dbbdab32ec054e)
	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			host := strings.ToLower(url.Host)

			if _, ok := betterttvDomains[host]; !ok {
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
