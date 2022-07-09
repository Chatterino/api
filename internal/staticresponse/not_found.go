package staticresponse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

var (
	NoLinkInfoFound = []byte(`{"status":404,"message":"Could not fetch link info: No link info found"}`)

	RNoLinkInfoFound = cache.Response{
		Payload:     NoLinkInfoFound,
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
	}

	SNoLinkInfoFound = &StaticResponse{
		payload:       NoLinkInfoFound,
		cacheDuration: resolver.NoSpecialDur,
	}

	NoThumbnailFound = []byte(`{"status":404,"message":"Could not fetch thumbnail"}`)

	SNoThumbnailFound = &StaticResponse{
		payload:       NoThumbnailFound,
		statusCode:    utils.IntPtr(http.StatusNotFound),
		cacheDuration: resolver.NoSpecialDur,
	}
)

func NotFoundf(format string, a ...interface{}) *StaticResponse {
	r := &resolver.Response{
		Status:  http.StatusNotFound,
		Message: resolver.CleanResponse(fmt.Sprintf(format, a...)),
	}

	payload, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return New(payload)
}
