package main

import (
	"encoding/json"
	"log"
	"os"
)

var (
	rNoLinkInfoFound []byte
	rInvalidURL      []byte
)

func init() {
	var err error
	r := &LinkResolverResponse{
		Status:  404,
		Message: "No link info found",
	}

	rNoLinkInfoFound, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}

	r = &LinkResolverResponse{
		Status:  500,
		Message: "Invalid URL",
	}
	rInvalidURL, err = json.Marshal(r)
	if err != nil {
		log.Println("Error marshalling prebuilt response:", err)
		os.Exit(1)
	}
}
