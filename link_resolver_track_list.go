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
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	invalidTrackPath = errors.New("invalid track list track path")
)

func init() {
	const (
		trackListAPIURL = "https://supinic.com/api/track/detail/%d"

		tooltipTemplate = `<div style="text-align: left;">
<b>{{.Name}}</b><br>
<b>Music Track</b><br>
<b>By:</b> {{.AuthorName}}<br>
<b>ID:</b> {{.ID}}<br>
<b>Duration:</b> {{.Duration}}<br>
<b>Tags:</b> {{.Tags}}</div>`
	)

	var (
		trackNotFoundResponse = &LinkResolverResponse{
			Status:  http.StatusNotFound,
			Message: "No track with this ID found",
		}
	)

	type TooltipData struct {
		ID         int
		Name       string
		AuthorName string
		Tags       string
		Duration   string
	}

	type TrackData struct {
		ID          int       `json:"id"`
		Link        string    `json:"code"` // Youtube ID/link
		Name        string    `json:"name"`
		VideoType   int       `json:"videoType"`
		TrackType   string    `json:"trackType"`
		Duration    int       `json:"duration"`
		Available   bool      `json:"available"`
		PublishedAt time.Time `json:"published"`
		Notes       string    `json:"notes"`
		AddedBy     string    `json:"addedBy"`
		ParsedLink  string    `json:"parsedLink"`
		Tags        []string  `json:"tags"`
		Authors     []struct {
			ID   int    `json:"ID"`
			Name string `json:"name"`
			Role string `json:"role"`
		} `json:"authors"`
	}

	type TrackListAPIResponse struct {
		Data TrackData `json:"data"`
	}

	tmpl, err := template.New("trackListEntryTooltip").Parse(tooltipTemplate)
	if err != nil {
		log.Println("Error initialization track list entry tooltip template:", err)
		return
	}

	load := func(rawTrackID string, r *http.Request) (interface{}, error, time.Duration) {
		trackID, _ := strconv.ParseInt(rawTrackID, 10, 32)
		apiURL := fmt.Sprintf(trackListAPIURL, trackID)

		// Create Track list API request
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Track list request creation error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

		// Execute Track list API request
		resp, err := httpClient.Do(req)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Track list http request error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		defer resp.Body.Close()

		// Error out if the track isn't found or something else went wrong with the request
		if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
			return trackNotFoundResponse, nil, noSpecialDur
		}

		// Read response into a string
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Track list http body read error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		// Parse response into a predefined JSON blob (see TrackListAPIResponse struct above)
		var jsonResponse TrackListAPIResponse
		if err := json.Unmarshal(body, &jsonResponse); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Track list api unmarshal error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		if jsonResponse.Data.ID == 0 { // API responds with
			return &LinkResolverResponse{
				Status:  http.StatusNotFound,
				Message: fmt.Sprintf("A track with ID %d doesn't exist.", trackID),
			}, nil, noSpecialDur
		}

		trackData := jsonResponse.Data

		prettyAuthors := ""
		for i, elem := range trackData.Authors {
			if i != 0 {
				prettyAuthors += ", "
			}
			prettyAuthors += fmt.Sprintf("%s (%d; %s)", elem.Name, elem.ID, elem.Role)

		}

		// formatDuration() cannot be use here as that parses an ISO duration
		duration := time.Duration(trackData.Duration) * time.Second
		hours := duration / time.Hour
		duration -= hours * time.Hour
		minutes := duration / time.Minute
		duration -= minutes * time.Minute
		seconds := duration / time.Second

		// Build tooltip data from the API response
		data := TooltipData{
			ID:         trackData.ID,
			Name:       trackData.Name,
			AuthorName: prettyAuthors,
			Tags:       strings.Join(trackData.Tags, ", "),
			Duration:   fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds),
		}

		// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
		var tooltip bytes.Buffer
		if err := tmpl.Execute(&tooltip, data); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Track list template error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:  200,
			Tooltip: tooltip.String(),
			//Thumbnail: thumbnailURL,
		}, nil, noSpecialDur
	}

	cache := newLoadingCache("tracklist_tracks", load, 1*time.Hour)
	trackPathRegex := regexp.MustCompile(`/track/detail/([0-9]+)`)

	// BetterTTV hosts we're doing our smart things on
	tracklistdomains := map[string]struct{}{
		"supinic.com": {},
	}

	// Find links matching the Track list link (e.g. https://supinic.com/track/detail/1883)
	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			host := strings.ToLower(url.Host)

			if _, ok := tracklistdomains[host]; !ok {
				return false
			}

			if !trackPathRegex.MatchString(url.Path) {
				return false
			}

			return true
		},
		run: func(url *url.URL) ([]byte, error) {
			matches := trackPathRegex.FindStringSubmatch(url.Path)
			if len(matches) != 2 {
				return nil, invalidTrackPath
			}

			trackID := matches[1]

			apiResponse := cache.Get(trackID, nil)
			return json.Marshal(apiResponse)
		},
	})
}
