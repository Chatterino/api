package discord

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

const discordInviteTooltip = `<div style="text-align: left;">
<b>{{.ServerName}}</b>
<br>
<br><b>Server Created:</b> {{.ServerCreated}}
<br><b>Channel:</b> {{.InviteChannel}}
{{ if .InviterTag}}<br><b>Inviter:</b> {{.InviterTag}}{{end}}
{{ if .ServerPerks}}<br><b>Server Perks:</b> {{.ServerPerks}}{{end}}
<br><b>Members:</b> <span style="color: #43b581;">{{.OnlineCount}} online</span>&nbsp;•&nbsp;<span style="color: #808892;">{{.TotalCount}} total</span>
</div>
`
