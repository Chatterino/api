package seventv

type EmoteAPIUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type EmoteAPIEmote struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Visibility int          `json:"visibility"`
	Owner      EmoteAPIUser `json:"owner"`
}

type EmoteAPIResponse struct {
	Data struct {
		Emote *EmoteAPIEmote `json:"emote,omitempty"`
	} `json:"data"`
}

type TooltipData struct {
	Code     string
	Type     string
	Uploader string
}
