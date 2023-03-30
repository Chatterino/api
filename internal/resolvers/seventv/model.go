package seventv

type TooltipData struct {
	Code     string
	Type     string
	Uploader string

	Unlisted bool
}

// Definitions from:
// * Emotes: https://github.com/SevenTV/API/blob/a907ccc44e7eb5bdba7b7e63d2b4b67e0c04f778/data/model/emote.model.go
// * Users: https://github.com/SevenTV/API/blob/a907ccc44e7eb5bdba7b7e63d2b4b67e0c04f778/data/model/user.model.go
// * Images: https://github.com/SevenTV/API/blob/a907ccc44e7eb5bdba7b7e63d2b4b67e0c04f778/data/model/model.go

type EmoteModel struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Flags  EmoteFlagsModel  `json:"flags"`
	Listed bool             `json:"listed"`
	Owner  UserPartialModel `json:"owner"`
	Host   ImageHost        `json:"host"`
}

type EmoteFlagsModel int32

const (
	// The emote is private and can only be accessed by its owner, editors and moderators
	EmoteFlagsPrivate EmoteFlagsModel = 1 << 0

	// The emote was verified to be an original creation by the uploader
	EmoteFlagsAuthentic EmoteFlagsModel = 1 << 1

	// The emote is recommended to be enabled as Zero-Width
	EmoteFlagsZeroWidth EmoteFlagsModel = 1 << 8

	// Content Flags

	// Sexually Suggesive
	EmoteFlagsContentSexual EmoteFlagsModel = 1 << 16

	// Rapid flashing
	EmoteFlagsContentEpilepsy EmoteFlagsModel = 1 << 17

	// Edgy or distasteful, may be offensive to some users
	EmoteFlagsContentEdgy EmoteFlagsModel = 1 << 18

	// Not allowed specifically on the Twitch platform
	EmoteFlagsContentTwitchDisallowed EmoteFlagsModel = 1 << 24
)

type ImageHost struct {
	URL   string      `json:"url"`
	Files []ImageFile `json:"files"`
}

type ImageFile struct {
	Name   string      `json:"name"`
	Width  int32       `json:"width"`
	Height int32       `json:"height"`
	Format ImageFormat `json:"format"`
}

type ImageFormat string

const (
	ImageFormatAVIF ImageFormat = "AVIF"
	ImageFormatWEBP ImageFormat = "WEBP"
)

type UserPartialModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}
