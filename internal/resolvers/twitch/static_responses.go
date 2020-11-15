package twitch

import (
	"net/http"

	"github.com/Chatterino/api/pkg/resolver"
)

var noTwitchClipWithThisIDFound = &resolver.Response{
	Status:  http.StatusNotFound,
	Message: "No Twitch Clip with this ID found",
}
