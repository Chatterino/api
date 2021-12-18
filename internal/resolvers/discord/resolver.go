package discord

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const discordInviteAPIURL = "https://discord.com/api/v9/invites/%s"

var (
	discordInviteURLRegex = regexp.MustCompile(`^(www\.)?discord\.(gg|com\/invite)\/([a-zA-Z0-9-]+)`)

	inviteCache = cache.New("discord_invites", load, 6*time.Hour) // Often calls quickly result in 429's

	inviteNotFoundResponse = &resolver.Response{
		Status:  http.StatusNotFound,
		Message: "No Discord invite with this code found",
	}

	errInvalidDiscordInvite = errors.New("invalid Discord invite Path")

	discordInviteTemplate = template.Must(template.New("discordInviteTooltip").Parse(discordInviteTooltip))

	token string
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	// Bot authentication is required for higher ratelimit (250 requests/5s)
	if cfg.DiscordToken == "" {
		log.Println("[Config] discord-token is missing, won't do special responses for Discord invites")
		return
	}
	token = cfg.DiscordToken

	// Find links matching the Discord invite link (e.g. https://discord.com/invite/mlp, https://discord.gg/mlp)
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
