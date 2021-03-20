package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"time"

	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/go-chi/chi/v5"
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	startTime = time.Now()

	bind    = flag.String("l", ":1234", "bind address")
	baseURL = flag.String("b", "", "base url (useful if being proxied through something like nginx). Value needs to be full url up to the application (e.g. https://braize.pajlada.com/chatterino)")

	prefix string
)

func mountRouter(r *chi.Mux) *chi.Mux {
	if *baseURL == "" {
		return r
	}

	// figure out prefix from address
	u, err := url.Parse(*baseURL)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Fatal("Scheme must be included in base url")
	}

	prefix = u.Path
	ur := chi.NewRouter()
	ur.Mount(prefix, r)

	log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", *bind, prefix, *baseURL)

	return ur
}

func listen(bind string, router *chi.Mux) {
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

	log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", *bind, prefix, *baseURL)

	router := chi.NewRouter()

	handleTwitchEmotes(router)
	handleHealth(router)

	defaultresolver.Initialize(router, *baseURL)
	defaultresolver.InitializeThumbnail(router)

	listen(*bind, mountRouter(router))
}
