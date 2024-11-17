package discord

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"regexp"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

const (
	discordInviteTooltip = `<div style="text-align: left;">
<b>{{.ServerName}}</b>
<br>
<br><b>Server Created:</b> {{.ServerCreated}}
<br><b>Channel:</b> {{.InviteChannel}}
{{ if .InviterTag}}<br><b>Inviter:</b> {{.InviterTag}}{{end}}
{{ if .ServerPerks}}<br><b>Server Perks:</b> {{.ServerPerks}}{{end}}
<br><b>Members:</b> <span style="color: #43b581;">{{.OnlineCount}} online</span>&nbsp;•&nbsp;<span style="color: #808892;">{{.TotalCount}} total</span>
</div>
`
)

var (
	discordInviteURLRegex = regexp.MustCompile(`^(www\.)?discord\.(gg|com\/invite)\/([a-zA-Z0-9-]+)`)

	inviteNotFoundResponse = &resolver.Response{
		Status:  http.StatusNotFound,
		Message: "No Discord invite with this code found",
	}

	errInvalidDiscordInvite = errors.New("invalid Discord invite Path")

	discordInviteTemplate = template.Must(template.New("discordInviteTooltip").Parse(discordInviteTooltip))
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)
	if cfg.DiscordToken == "" {
		log.Warnw("[Config] discord-token is missing, won't do special responses for Discord invites")
		return
	}

	apiURL := utils.MustParseURL("https://discord.com/api/v9/invites/")

	*resolvers = append(*resolvers, NewInviteResolver(ctx, cfg, pool, apiURL))
}
