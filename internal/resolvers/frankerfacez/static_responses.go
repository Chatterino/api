package frankerfacez

import (
	"net/http"

	"github.com/Chatterino/api/pkg/resolver"
)

var (
	emoteNotFoundResponse = &resolver.Response{
		Status:  http.StatusNotFound,
		Message: "No FrankerFaceZ emote with this id found",
	}
)
