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
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func init() {
	const (
		discordInviteAPIURL = "https://discord.com/api/v6/invites/%s?with_counts=true"

		discordInviteTooltip = `<div style="text-align: left;">
<b>{{.ServerName}}</b>
<br>
<br><b>Server Created:</b> {{.ServerCreated}}
<br><b>Channel:</b> {{.InviteChannel}}
{{.InviterTag}}
{{.ServerPerks}}
<br><b>Members:</b> <span style="color: #43b581;">{{.OnlineCount}} online</span>&nbsp;•&nbsp;<span style="color: #808892;">{{.TotalCount}} total</span>
</div>
`
	)

	var (
		inviteNotFoundResponse = &LinkResolverResponse{
			Status:  http.StatusNotFound,
			Message: "No Discord invite with this code found",
		}

		invalidDiscordInvite = errors.New("Invalid Discord invite Path")
	)

	type TooltipData struct {
		ServerName    string
		ServerCreated string
		InviteChannel string
		InviterTag    string
		ServerPerks   string
		OnlineCount   string
		TotalCount    string
	}

	type DiscordInviteData struct {
		Message string `json:"message,omitempty"`
		Guild   struct {
			ID       string   `json:"id"`
			Name     string   `json:"name"`
			IconHash string   `json:"icon"`
			Features []string `json:"features"`
		} `json:"guild,omitempty"`
		Channel struct {
			Name string `json:"name"`
		} `json:"channel,omitempty"`
		Inviter struct {
			Username      string `json:"username"`
			Discriminator string `json:"discriminator"`
		} `json:"inviter,omitempty"`
		OnlineCount int64 `json:"approximate_presence_count,omitempty"`
		TotalCount  int64 `json:"approximate_member_count,omitempty"`
	}

	// Bot authentication is required for higher ratelimit (250 requests/5s)
	discordToken, exists := os.LookupEnv("CHATTERINO_API_DISCORD_TOKEN")
	if !exists {
		log.Println("No CHATTERINO_API_DISCORD_TOKEN specified, won't do special responses for Discord invites")
		return
	}

	tmpl, err := template.New("discordInviteTooltip").Parse(discordInviteTooltip)
	if err != nil {
		log.Println("Error while initializing Discord invite tooltip template:", err)
		return
	}

	load := func(inviteCode string, r *http.Request) (interface{}, error, time.Duration) {
		log.Println("[DiscordInvite] GET", inviteCode)
		apiURL := fmt.Sprintf(discordInviteAPIURL, inviteCode)

		// Create Discord API request
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Discord API request creation error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")
		req.Header.Set("Authorization", fmt.Sprintf("Bot %s", discordToken))

		// Execute Discord API request
		resp, err := httpClient.Do(req)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Discord API request error " + clean(err.Error()),
			}, nil, noSpecialDur
		}
		defer resp.Body.Close()

		// Error out if the invite isn't found or something else went wrong with the request
		if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
			return inviteNotFoundResponse, nil, noSpecialDur
		}

		// Read response into a string
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Discord API http body read error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		// Parse response into a predefined JSON blob (see TrackListAPIResponse struct above)
		var jsonResponse DiscordInviteData
		if err := json.Unmarshal(body, &jsonResponse); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Discord API unmarshal error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		// API doesn't include "approximate_member_count" if an invite was not found
		if jsonResponse.TotalCount == 0 {
			return inviteNotFoundResponse, nil, noSpecialDur
		}

		// Some dank utils (decided to keep those here, as they will be useless outside this file)
		// Comverting Discord Snowflake to date string
		// Reference https://discord.com/developers/docs/reference#snowflakes
		snowflake, _ := strconv.ParseInt(jsonResponse.Guild.ID, 10, 64)
		getDateFromSnowflake := time.Unix(snowflake>>22/1000+1420066800, 0).Format("02 Jan 2006")

		// Adding row with inviter's user tag if present
		getInviter := ""
		if jsonResponse.Inviter.Username != "" {
			userTag := fmt.Sprintf("%s#%s", jsonResponse.Inviter.Username, jsonResponse.Inviter.Discriminator)
			getInviter = fmt.Sprintf("<br><b>Inviter:</b> %s", userTag)
		}

		// Parsing only selected meaningful server perks
		parsePerks := ""
		accpetedPerks := []string{"PARTNERED", "PUBLIC", "ANIMATED_ICON", "BANNER", "INVITE_SPLASH", "VIP_REGIONS", "VANITY_URL"}
		for _, elem := range jsonResponse.Guild.Features {
			if contains(accpetedPerks, elem) {
				if parsePerks != "" {
					parsePerks += ", "
				}
				parsePerks += strings.ToLower(strings.Replace(elem, "_", " ", -1))
			}
		}
		if parsePerks != "" {
			parsePerks = fmt.Sprintf("<br><b>Server Perks:</b> %s", parsePerks)
		}

		// Build tooltip data from the API response
		data := TooltipData{
			ServerName:    jsonResponse.Guild.Name,
			ServerCreated: getDateFromSnowflake,
			InviteChannel: fmt.Sprintf("#%s", jsonResponse.Channel.Name),
			InviterTag:    getInviter,
			ServerPerks:   parsePerks,
			OnlineCount:   insertCommas(strconv.FormatInt(jsonResponse.OnlineCount, 10), 3),
			TotalCount:    insertCommas(strconv.FormatInt(jsonResponse.TotalCount, 10), 3),
		}

		// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
		var tooltip bytes.Buffer
		if err := tmpl.Execute(&tooltip, data); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Discord Invite template error " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:    200,
			Tooltip:   url.PathEscape(tooltip.String()),
			Thumbnail: fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s", jsonResponse.Guild.ID, jsonResponse.Guild.IconHash),
			Link:      fmt.Sprintf("https://discord.gg/%s", inviteCode),
		}, nil, noSpecialDur
	}

	cache := newLoadingCache("discord_invites", load, 6*time.Hour) // Often calls quickly result in 429's
	discordInviteURLRegex := regexp.MustCompile(`^(www\.)?(discord\.gg|discord(app)?\.com\/invite)\/([a-zA-Z0-9-]+)`)

	// Find links matching the Discord invite link (e.g. https://discord.com/invite/mlp, https://discord.gg/mlp)
	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			return discordInviteURLRegex.MatchString(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
		},
		run: func(url *url.URL) ([]byte, error) {
			matches := discordInviteURLRegex.FindStringSubmatch(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
			if len(matches) != 5 {
				return nil, invalidDiscordInvite
			}

			inviteCode := matches[4]

			apiResponse := cache.Get(inviteCode, nil)
			return json.Marshal(apiResponse)
		},
	})
}
