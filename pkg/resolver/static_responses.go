package resolver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	NoLinkInfoFound  []byte
	InvalidURL       []byte
	ResponseTooLarge []byte
)

func init() {
	var err error
	r := &Response{
		Status:  404,
		Message: "Could not fetch link info: No link info found",
	}

	NoLinkInfoFound, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}

	r = &Response{
		Status:  500,
		Message: "Could not fetch link info: Invalid URL",
	}
	InvalidURL, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}

	r = &Response{
		Status:  http.StatusInternalServerError,
		Message: fmt.Sprintf("Could not fetch link info: Response too large (>%dMB)", MaxContentLength/1024/1024),
	}
	ResponseTooLarge, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}
}
