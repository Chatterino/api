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
