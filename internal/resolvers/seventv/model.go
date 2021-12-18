package seventv

type EmoteAPIUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type EmoteAPIEmote struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Visibility int32        `json:"visibility"`
	Owner      EmoteAPIUser `json:"owner"`
}

type EmoteAPIResponseData struct {
	Emote *EmoteAPIEmote `json:"emote,omitempty"`
}

type EmoteAPIResponse struct {
	Data EmoteAPIResponseData `json:"data"`
}

type TooltipData struct {
	Code     string
	Type     string
	Uploader string

	Unlisted bool
}

const (
	EmoteVisibilityPrivate int32 = 1 << iota
	EmoteVisibilityGlobal
	EmoteVisibilityHidden
	EmoteVisibilityOverrideBTTV
	EmoteVisibilityOverrideFFZ
	EmoteVisibilityOverrideTwitchGlobal
	EmoteVisibilityOverrideTwitchSubscriber

	EmoteVisibilityAll int32 = (1 << iota) - 1
)

const tooltipTemplate = `<div style="text-align: left;">
<b>{{.Code}}</b><br>
<b>{{.Type}} SevenTV Emote</b><br>
<b>By:</b> {{.Uploader}}` +
	`{{ if .Unlisted }}` + `
<li><b><span style="color: red;">UNLISTED</span></b></li>{{ end }}
</div>`
