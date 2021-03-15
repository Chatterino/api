package livestreamfails

import (
	"net/http"

	"github.com/Chatterino/api/pkg/resolver"
)

var noLivestreamfailsClipWithThisIDFound = &resolver.Response{
	Status:  http.StatusNotFound,
	Message: "No LivestreamFails Clip with this ID found",
}
