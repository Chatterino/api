package twitch

const goodSlugV1 = "GoodSlugV1"
const goodSlugV2 = "GoodSlugV2-HVUvT7bYQnMn6nwp"
const goodSlugV3 = "EndearingPhilanthropicLEDDAESuppy"

var validClipBase = []string{
	"https://clips.twitch.tv/",
	"https://twitch.tv/pajlada/clip/",
	"https://twitch.tv/zneix/clip/",
	"https://m.twitch.tv/pajlada/clip/",
	"https://m.twitch.tv/zneix/clip/",
	"https://m.twitch.tv/clip/",
	"https://m.twitch.tv/clip/clip/",
}

// clips that are invalid due to path or domain+path combination
var invalidClips = []string{
	"https://clips.twitch.tv/pajlada/clip/VastBitterVultureMau5",
	"https://clips.twitch.tv/",
	"https://twitch.tv/nam____________________________________________/clip/someSlugNam",
	"https://twitch.tv/supinic/clip/",
	"https://twitch.tv/pajlada/clips/VastBitterVultureMau5",
	"https://twitch.tv/zneix/clip/ImpossibleOilyAlpacaTF2John-jIlgtnSAQ52BThHhifyouseethisvivon",
	"https://twitch.tv/clip/slug",
	"https://gql.twitch.tv/VastBitterVultureMau5",
	"https://gql.twitch.tv/ThreeLetterAPI/clip/VastBitterVultureMau5",
	"https://m.twitch.tv/VastBitterVultureMau5",
	"https://m.twitch.tv/username/clip/clip/slug",
	"https://m.twitch.tv/username/notclip/slug",
}

// clips that are invalid due to the path
var invalidClipSlugs = []string{
	"https://clips.twitch.tv/",
	"https://twitch.tv/nam____________________________________________/clip/someSlugNam",
	"https://twitch.tv/supinic/clip/",
	"https://twitch.tv/pajlada/clips/VastBitterVultureMau5",
	"https://twitch.tv/zneix/clip/ImpossibleOilyAlpacaTF2John-jIlgtnSAQ52BThHhifyouseethisvivon",
	"https://m.twitch.tv/username/notclip/slug",
}

var validUsers = []string{
	"https://twitch.tv/pajlada",
	"https://www.twitch.tv/pajlada",
	"https://twitch.tv/matthewde",
}

var invalidUsers = []string{
	"https://twitch.tv/inventory",
	"https://twitch.tv/popout",
	"https://twitch.tv/subscriptions",
	"https://twitch.tv/videos",
	"https://twitch.tv/following",
	"https://twitch.tv/directory",
	"https://twitch.tv/DIRECTORY",
	"https://twitch.tv/moderator",
	"https://twitch.com/matthewde",
	"https://clips.twitch.tv/EndearingPhilanthropicLEDDAESuppy",
}
