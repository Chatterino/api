package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
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
	OnlineCount uint64 `json:"approximate_presence_count,omitempty"`
	TotalCount  uint64 `json:"approximate_member_count,omitempty"`
}

type InviteLoader struct {
	baseURL *url.URL

	token string
}

func NewInviteLoader(baseURL *url.URL, token string) *InviteLoader {
	l := &InviteLoader{
		baseURL: baseURL,

		token: token,
	}

	return l
}

func (l *InviteLoader) buildURL(inviteCode string) *url.URL {
	relativeURL := &url.URL{
		Path: inviteCode,
	}
	finalURL := l.baseURL.ResolveReference(relativeURL)

	return finalURL
}

func (l *InviteLoader) Load(ctx context.Context, inviteCode string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)
	log.Debugw("[DiscordInvite] Get invite",
		"inviteCode", inviteCode,
	)

	apiURL := l.buildURL(inviteCode)
	apiURLVariables := url.Values{}
	apiURLVariables.Set("with_counts", "true")
	apiURL.RawQuery = apiURLVariables.Encode()

	extraHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bot %s", l.token),
	}

	// Execute Discord API request
	resp, err := resolver.RequestGETWithHeaders(apiURL.String(), extraHeaders)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Discord API request error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}
	defer resp.Body.Close()

	// Error out if the invite isn't found or something else went wrong with the request
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return inviteNotFoundResponse, cache.NoSpecialDur, nil
	}

	// Read response into a string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Discord API http body read error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// Parse response into a predefined JSON blob (see TrackListAPIResponse struct above)
	var jsonResponse DiscordInviteData
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Discord API unmarshal error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	// API doesn't include "approximate_member_count" if an invite was not found
	if jsonResponse.TotalCount == 0 {
		return inviteNotFoundResponse, cache.NoSpecialDur, nil
	}

	// Some dank utils (decided to keep those here, as they will be useless outside this file)
	// Comverting Discord Snowflake to date string
	// Reference https://discord.com/developers/docs/reference#snowflakes
	snowflake, _ := strconv.ParseInt(jsonResponse.Guild.ID, 10, 64)
	dateFromSnowflake := humanize.CreationDateUnix(snowflake>>22/1000 + 1420066800)

	// Adding row with inviter's user tag if present
	userTag := ""
	if jsonResponse.Inviter.Username != "" {
		userTag = fmt.Sprintf("%s#%s", jsonResponse.Inviter.Username, jsonResponse.Inviter.Discriminator)
	}

	// Parsing only selected meaningful server perks
	// An example of a server that has pretty much all the perks: https://discord.com/api/invites/test
	parsedPerks := ""
	accpetedPerks := []string{"PARTNERED", "PUBLIC", "ANIMATED_ICON", "BANNER", "INVITE_SPLASH", "VIP_REGIONS", "VANITY_URL", "COMMUNITY"}
	slices.SortStableFunc(jsonResponse.Guild.Features, func(a, b string) int {
		return strings.Compare(a, b)
	})
	for _, elem := range jsonResponse.Guild.Features {
		if utils.Contains(accpetedPerks, elem) {
			if parsedPerks != "" {
				parsedPerks += ", "
			}
			parsedPerks += strings.ToLower(strings.Replace(elem, "_", " ", -1))
		}
	}

	// Build tooltip data from the API response
	data := TooltipData{
		ServerName:    jsonResponse.Guild.Name,
		ServerCreated: dateFromSnowflake,
		InviteChannel: fmt.Sprintf("#%s", jsonResponse.Channel.Name),
		InviterTag:    userTag,
		ServerPerks:   parsedPerks,
		OnlineCount:   humanize.Number(jsonResponse.OnlineCount),
		TotalCount:    humanize.Number(jsonResponse.TotalCount),
	}

	// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
	var tooltip bytes.Buffer
	if err := discordInviteTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Discord Invite template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s", jsonResponse.Guild.ID, jsonResponse.Guild.IconHash),
		Link:      fmt.Sprintf("https://discord.gg/%s", inviteCode),
	}, cache.NoSpecialDur, nil

}
