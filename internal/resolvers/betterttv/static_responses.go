package betterttv

import (
	"net/http"

	"github.com/Chatterino/api/pkg/resolver"
)

var (
	emoteNotFoundResponse = &resolver.Response{
		Status:  http.StatusNotFound,
		Message: "No BetterTTV emote with this hash found",
	}
)
