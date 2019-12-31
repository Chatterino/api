package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	maxContentLength = 1024 * 1024 * 5 // 5MB
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	startTime = time.Now()
)

var host = flag.String("h", ":1234", "host of server")
var prefix = flag.String("p", "", "optional prefix")

func makeRouter(prefix string) *mux.Router {
	// Skip clean is used to make link_resolver work
	router := mux.NewRouter().SkipClean(true)
	sr := router.PathPrefix(prefix).Subrouter().SkipClean(true)

	return sr
}

func listen(host string, router *mux.Router) {
	srv := &http.Server{
		Handler:      router,
		Addr:         host,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Listening on", host)
	log.Fatal(srv.ListenAndServe())
}

func main() {
	flag.Parse()

	router := makeRouter(*prefix)

	handleTwitchEmotes(router)

	handleHealth(router)

	handleLinkResolver(router)

	listen(*host, router)
}
