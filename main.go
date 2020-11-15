package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"time"

	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/gorilla/mux"
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	startTime = time.Now()
)

var bind = flag.String("l", ":1234", "bind address")
var baseURL = flag.String("b", "", "base url (useful if being proxied through nginx or some shit). value needs to be full url up to the application (e.g. https://braize.pajlada.com/chatterino)")

var prefix string

func makeRouter(prefix string) *mux.Router {
	// Skip clean is used to make link_resolver work
	router := mux.NewRouter().SkipClean(true)
	sr := router.PathPrefix(prefix).Subrouter().SkipClean(true)

	return sr
}

func listen(bind string, router *mux.Router) {
	srv := &http.Server{
		Handler:      router,
		Addr:         bind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func main() {
	flag.Parse()

	// figure out prefix from baseURL
	if *baseURL != "" {
		u, err := url.Parse(*baseURL)
		if err != nil {
			log.Fatal(err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			log.Fatal("scheme must be included in base url")
		}
		prefix = u.Path
	}

	log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", *bind, prefix, *baseURL)

	router := makeRouter(prefix)

	handleTwitchEmotes(router)

	handleHealth(router)

	defaultresolver.Initialize(router, *baseURL)

	defaultresolver.InitializeThumbnail(router)

	listen(*bind, router)
}
