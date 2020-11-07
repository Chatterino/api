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
	errInvalidFFZEmotePath = errors.New("invalid FrankerFaceZ emote path")
)

func init() {
	const (
		emoteAPIURL = "https://api.frankerfacez.com/v1/emote/%s"

		thumbnailFormat = "https://cdn.frankerfacez.com/emoticon/%s/4"

		tooltipTemplate = `<div style="text-align: left;">
<b>{{.Code}}</b><br>
<b>FrankerFaceZ Emote</b><br>
<b>By:</b> {{.Uploader}}</div>`
	)

	var (
		emoteNotFoundResponse = &LinkResolverResponse{
			Status:  http.StatusNotFound,
			Message: "No FrankerFaceZ emote with this id found",
		}
	)

	type TooltipData struct {
		Code     string
		Uploader string
	}

	/* Example JSON data generated from https://api.frankerfacez.com/v1/emote/131001 2020-11-07
	{
	  "emote": {
	    "created_at": "Sun, 25 Sep 2016 12:30:30 GMT",
	    "css": null,
	    "height": 21,
	    "hidden": false,
	    "id": 131001,
	    "last_updated": "Sun, 25 Sep 2016 14:25:01 GMT",
	    "margins": null,
	    "modifier": false,
	    "name": "pajaE",
	    "offset": null,
	    "owner": {
	      "_id": 63119,
	      "display_name": "pajaSWA",
	      "name": "pajaswa"
	    },
	    "public": true,
	    "status": 1,
	    "urls": {
	      "1": "//cdn.frankerfacez.com/8542ab940f02f3bdc938796dc7258902.PNG",
	      "2": "//cdn.frankerfacez.com/c1fe2e20b1d13e97b40b44f6893a7ba4.PNG",
	      "4": "//cdn.frankerfacez.com/6154d1c0f922ee6506cb2e555dd46e03.png"
	    },
	    "usage_count": 9,
	    "width": 32
	  }
	}
	*/
	type FrankerFaceZEmoteAPIResponse struct {
		Height       int16  `json:"height"`
		Modifier     bool   `json:"modifier"`
		Status       int    `json:"status"`
		Width        int16  `json:"width"`
		Hidden       bool   `json:"hidden"`
		CreatedAtRaw string `json:"created_at"`
		CreatedAt    time.Time
		UpdatedAtRaw string `json:"last_updated"`
		UpdatedAt    time.Time
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Public       bool   `json:"public"`
		Owner        struct {
			DisplayName string `json:"display_name"`
			ID          int    `json:"_id"`
			Name        string `json:"name"`
		} `json:"owner"`

		URLs struct {
			Size1 string `json:"1"`
			Size2 string `json:"2"`
			Size4 string `json:"4"`
		} `json:"urls"`
	}

	tmpl, err := template.New("frankerfacezEmoteTooltip").Parse(tooltipTemplate)
	if err != nil {
		log.Println("Error initialization FFZ Emotes tooltip template:", err)
		return
	}

	load := func(emoteID string, r *http.Request) (interface{}, error, time.Duration) {
		apiURL := fmt.Sprintf(emoteAPIURL, emoteID)
		thumbnailURL := fmt.Sprintf(thumbnailFormat, emoteID)

		// Create FFZ API request
		resp, err := makeRequest(apiURL)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "ffz http request error " + clean(err.Error()),
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
				Message: "ffz http body read error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		// Parse response into a predefined JSON blob (see FrankerFaceZEmoteAPIResponse struct above)
		var temp struct {
			Emote FrankerFaceZEmoteAPIResponse `json:"emote"`
		}

		if err := json.Unmarshal(body, &temp); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "ffz api unmarshal error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		jsonResponse := temp.Emote
		jsonResponse.CreatedAt, err = time.Parse(time.RFC1123, jsonResponse.CreatedAtRaw)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "ffz api created at time unmarshal error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		jsonResponse.UpdatedAt, err = time.Parse(time.RFC1123, jsonResponse.UpdatedAtRaw)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "ffz api updated at time unmarshal error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		// Build tooltip data from the API response
		data := TooltipData{
			Code:     jsonResponse.Name,
			Uploader: jsonResponse.Owner.DisplayName,
		}

		// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
		var tooltip bytes.Buffer
		if err := tmpl.Execute(&tooltip, data); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "ffz template error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:    200,
			Tooltip:   url.PathEscape(tooltip.String()),
			Thumbnail: thumbnailURL,
		}, nil, noSpecialDur
	}

	cache := newLoadingCache("ffz_emotes", load, 1*time.Hour)
	emotePathRegex := regexp.MustCompile(`/emoticon/([0-9]+)-.+?`)

	// FFZ hosts we're doing our smart things on
	ffzDomains := map[string]struct{}{
		"frankerfacez.com":     {},
		"www.frankerfacez.com": {},
	}

	// Find links matching the FFZ direct emote link (e.g. https://www.frankerfacez.com/emoticon/490944-PAJAP)
	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			host := strings.ToLower(url.Host)

			if _, ok := ffzDomains[host]; !ok {
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
				return nil, errInvalidFFZEmotePath
			}

			emoteHash := matches[1]

			apiResponse := cache.Get(emoteHash, nil)
			return json.Marshal(apiResponse)
		},
	})
}
