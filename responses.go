package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	rNoLinkInfoFound  []byte
	rInvalidURL       []byte
	rResponseTooLarge []byte
)

func init() {
	var err error
	r := &LinkResolverResponse{
		Status:  404,
		Message: "Could not fetch link info: No link info found",
	}

	rNoLinkInfoFound, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}

	r = &LinkResolverResponse{
		Status:  500,
		Message: "Could not fetch link info: Invalid URL",
	}
	rInvalidURL, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}

	r = &LinkResolverResponse{
		Status:  http.StatusInternalServerError,
		Message: fmt.Sprintf("Could not fetch link info: Response too large (>%dMB)", maxContentLength/1024/1024),
	}
	rResponseTooLarge, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}
}
