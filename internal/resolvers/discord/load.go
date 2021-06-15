package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

func load(inviteCode string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[DiscordInvite] GET", inviteCode)
	apiURL := fmt.Sprintf(discordInviteAPIURL, inviteCode)
	extraHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bot %s", token),
	}

	// Execute Discord API request
	resp, err := resolver.RequestGETWithHeaders(apiURL, extraHeaders)
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
	body, err := ioutil.ReadAll(resp.Body)
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
