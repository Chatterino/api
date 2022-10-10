package staticresponse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Chatterino/api/pkg/resolver"
)

func InternalServerErrorf(format string, a ...interface{}) *StaticResponse {
	r := &resolver.Response{
		Status:  http.StatusInternalServerError,
		Message: resolver.CleanResponse(fmt.Sprintf(format, a...)),
	}

	payload, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return New(payload)
}
